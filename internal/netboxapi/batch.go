package netboxapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// Terraform reads every resource on its own, ten at a time by default. NetBox spends most of the
// time of a small request on work that is the same for every request (the token, the permissions,
// the transaction), so ten requests for one object each take several times as long as one request
// for ten objects. The client therefore gathers the reads of single objects that wait at the same
// time and asks for them in one list request, filtered by ID.
//
// No time is spent waiting for more reads to arrive. The first read of a collection is sent at
// once. The reads that come in while a request for that collection is under way are sent together
// when it has finished. A single read costs what it did before, and the busier NetBox is, the
// larger the batches get.
//
// A list returns the same fields for an object as reading it alone. When an object is not in the
// answer, or the list request fails, the object is read alone, so a batch never decides that an
// object is gone or that a read failed.

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

	// busy tells that a request for the collection is under way.
	busy bool
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
	if !g.busy {
		// Nothing is under way: send the read now. What arrives meanwhile goes out afterwards.
		g.busy = true
		c.batchMu.Unlock()
		raw, err := c.getOne(ctx, path, id, query)
		go c.drain(g)
		return raw, err
	}
	call := &getCall{ctx: ctx, id: id, done: make(chan struct{})}
	g.waiting = append(g.waiting, call)
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

// drain sends the waiting reads of the group, batch after batch, until none is left.
func (c *Client) drain(g *batchGroup) {
	for {
		c.batchMu.Lock()
		if len(g.waiting) == 0 {
			g.busy = false
			c.batchMu.Unlock()
			return
		}
		n := min(len(g.waiting), maxBatch)
		calls := g.waiting[:n:n]
		g.waiting = g.waiting[n:]
		c.batchMu.Unlock()

		if !c.fetchBatch(g, calls) {
			c.batchMu.Lock()
			g.single = true
			c.batchMu.Unlock()
		}
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
