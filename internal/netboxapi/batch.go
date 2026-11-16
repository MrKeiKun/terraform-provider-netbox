package netboxapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Terraform reads every resource on its own, ten at a time by default. NetBox spends most of the
// time of a small request on work that is the same for every request (the token, the permissions,
// the transaction), so ten requests for one object each take several times as long as one request
// for ten objects. The client therefore gathers the reads of single objects that arrive close
// together and asks for them in one list request, filtered by ID.
//
// A read waits for at most gatherWindow, then goes out with the reads of the same collection that
// arrived meanwhile. Batches do not wait for each other. Holding reads back until the request before
// them has finished makes larger batches, but it turned out slower than no batching at all as soon
// as NetBox takes 50 ms or more to answer, because only one request is then under way at a time.
//
// A list returns the same fields for an object as reading it alone. When an object is not in the
// answer, or the list request fails, the object is read alone, so a batch never decides that an
// object is gone or that a read failed.

// gatherWindow is how long a read waits for others to share a request with. Terraform asks for the
// next resources within a few milliseconds after the ones before them are done.
var gatherWindow = 5 * time.Millisecond

// maxBatch is the most objects asked for in one request. It stays far below the page size NetBox
// allows, and keeps the URL short.
const maxBatch = 50

// SetBatchReads turns the gathering of single reads into list requests on or off.
func (c *Client) SetBatchReads(on bool) {
	c.batchReads = on
}

// getCall is one read of an object that waits for a batch.
type getCall struct {
	ctx  context.Context
	id   int64
	raw  json.RawMessage
	err  error
	done chan struct{}
}

// batchGroup holds the reads of one collection, with the same query parameters.
type batchGroup struct {
	path  string
	query url.Values

	// single tells that the collection is read object by object, because a batch did not work.
	single  bool
	waiting []*getCall
}

// getRaw reads one object and returns its JSON.
func (c *Client) getRaw(ctx context.Context, path string, id int64, query url.Values) (json.RawMessage, error) {
	if !c.batchReads {
		return c.getOne(ctx, path, id, query)
	}

	key := path + "?" + query.Encode()
	c.batchMu.Lock()
	if c.batches == nil {
		c.batches = map[string]*batchGroup{}
	}
	g := c.batches[key]
	if g == nil {
		g = &batchGroup{path: path, query: query}
		c.batches[key] = g
	}
	if g.single {
		c.batchMu.Unlock()
		return c.getOne(ctx, path, id, query)
	}
	call := &getCall{ctx: ctx, id: id, done: make(chan struct{})}
	g.waiting = append(g.waiting, call)
	switch {
	case len(g.waiting) == 1:
		// The first read of a batch starts the clock.
		time.AfterFunc(gatherWindow, func() { c.flush(g) })
	case len(g.waiting) >= maxBatch:
		go c.flush(g)
	}
	c.batchMu.Unlock()

	select {
	case <-call.done:
		return call.raw, call.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// getOne reads one object with a request of its own.
func (c *Client) getOne(ctx context.Context, path string, id int64, query url.Values) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.Do(ctx, http.MethodGet, c.endpoint(objectPath(path, id), query), nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// flush sends the reads that wait for the group.
func (c *Client) flush(g *batchGroup) {
	c.batchMu.Lock()
	n := min(len(g.waiting), maxBatch)
	calls := g.waiting[:n:n]
	g.waiting = g.waiting[n:]
	if len(g.waiting) > 0 {
		go c.flush(g)
	}
	c.batchMu.Unlock()
	if n == 0 {
		return
	}
	if !c.fetchBatch(g, calls) {
		c.batchMu.Lock()
		g.single = true
		c.batchMu.Unlock()
	}
}

// fetchBatch reads the objects of the calls with one list request and hands every call its object.
// It reports false when the list did not serve the calls, so that the collection is read object by
// object from then on.
func (c *Client) fetchBatch(g *batchGroup, calls []*getCall) bool {
	// A read that was given up no longer needs an answer.
	live := calls[:0:0]
	for _, call := range calls {
		if call.ctx.Err() == nil {
			live = append(live, call)
		}
	}
	if len(live) == 0 {
		return true
	}
	if len(live) == 1 {
		c.finishAlone(g, live)
		return true
	}

	query := url.Values{}
	for k, v := range g.query {
		query[k] = v
	}
	seen := map[int64]bool{}
	for _, call := range live {
		if !seen[call.id] {
			seen[call.id] = true
			query.Add("id", strconv.FormatInt(call.id, 10))
		}
	}
	query.Set("limit", strconv.Itoa(len(seen)))

	// The request serves several callers, so it must not stop when one of them gives up. The
	// timeout of the HTTP client still ends it.
	ctx := context.WithoutCancel(live[0].ctx)
	page, err := ListPage[json.RawMessage](ctx, c, g.path, query)
	if err != nil {
		// Reading the objects alone gives every caller the error that belongs to its object.
		c.finishAlone(g, live)
		return false
	}

	found := make(map[int64]json.RawMessage, len(page.Results))
	for _, raw := range page.Results {
		var head struct {
			ID *int64 `json:"id"`
		}
		if json.Unmarshal(raw, &head) == nil && head.ID != nil && seen[*head.ID] {
			found[*head.ID] = raw
		}
	}

	ok := true
	var missing []*getCall
	for _, call := range live {
		if raw, hit := found[call.id]; hit {
			call.raw = raw
			close(call.done)
		} else {
			missing = append(missing, call)
		}
	}
	if len(missing) > 0 {
		// An object that is not in the list is usually gone. Reading it alone tells for sure. When it
		// is there after all, the list cannot be relied on for this collection.
		c.finishAlone(g, missing)
		for _, call := range missing {
			if call.err == nil {
				ok = false
			}
		}
	}
	return ok
}

// finishAlone reads the object of every call with its own request, side by side.
func (c *Client) finishAlone(g *batchGroup, calls []*getCall) {
	var wg sync.WaitGroup
	for _, call := range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			call.raw, call.err = c.getOne(call.ctx, g.path, call.id, g.query)
			close(call.done)
		}()
	}
	wg.Wait()
}
