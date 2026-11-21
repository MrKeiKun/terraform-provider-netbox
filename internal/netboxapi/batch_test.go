package netboxapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// batchServer is a NetBox that has the objects 1 to 100 of one collection. It counts the requests
// for a single object and the list requests.
type batchServer struct {
	*httptest.Server
	singles, lists atomic.Int32
	// gone are IDs that do not exist.
	gone map[int64]bool
	// listStatus is the answer to list requests when it is not zero.
	listStatus int
	// ignoreFilter makes list requests return the objects 1000 and up, whatever was asked for.
	ignoreFilter bool
	// wait holds back every answer, so that reads pile up.
	wait time.Duration
}

func newBatchServer(t *testing.T) *batchServer {
	s := &batchServer{gone: map[int64]bool{}, wait: 10 * time.Millisecond}
	object := func(id int64) string { return fmt.Sprintf(`{"id": %d, "name": "object %d"}`, id, id) }
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(s.wait)
		rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/dcim/sites"), "/")
		if rest != "" {
			s.singles.Add(1)
			id, _ := strconv.ParseInt(rest, 10, 64)
			if s.gone[id] {
				http.Error(w, `{"detail": "Not found."}`, http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(object(id)))
			return
		}
		s.lists.Add(1)
		if s.listStatus != 0 {
			http.Error(w, `{"detail": "no lists here"}`, s.listStatus)
			return
		}
		var objs []string
		for _, raw := range r.URL.Query()["id"] {
			id, _ := strconv.ParseInt(raw, 10, 64)
			if s.ignoreFilter {
				id += 1000
			}
			if !s.gone[id] {
				objs = append(objs, object(id))
			}
		}
		if limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); limit < len(objs) {
			t.Errorf("the limit %d is lower than the %d objects asked for", limit, len(objs))
		}
		_, _ = fmt.Fprintf(w, `{"count": %d, "next": null, "results": [%s]}`, len(objs), strings.Join(objs, ","))
	}))
	t.Cleanup(s.Close)
	return s
}

func (s *batchServer) client(t *testing.T, batch bool) *Client {
	c, err := New(s.URL, "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	c.SetMaxRetries(0)
	c.SetBatchReads(batch)
	return c
}

type named struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// readAll reads the objects with the given IDs side by side and returns the error of each read.
func readAll(t *testing.T, c *Client, ids []int64) map[int64]error {
	var mu sync.Mutex
	errs := map[int64]error{}
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obj, err := Get[named](context.Background(), c, "dcim/sites/", id)
			if err == nil && (obj.ID != id || obj.Name != fmt.Sprintf("object %d", id)) {
				err = fmt.Errorf("read %d, got %+v", id, obj)
			}
			mu.Lock()
			errs[id] = err
			mu.Unlock()
		}()
	}
	wg.Wait()
	return errs
}

func ids(from, to int64) []int64 {
	var out []int64
	for i := from; i <= to; i++ {
		out = append(out, i)
	}
	return out
}

func TestReadsAtTheSameTimeShareARequest(t *testing.T) {
	s := newBatchServer(t)
	c := s.client(t, true)
	for id, err := range readAll(t, c, ids(1, 40)) {
		if err != nil {
			t.Errorf("object %d: %v", id, err)
		}
	}
	requests := s.singles.Load() + s.lists.Load()
	if requests > 10 {
		t.Errorf("40 reads took %d requests (%d single, %d lists)", requests, s.singles.Load(), s.lists.Load())
	}
	// With nothing else under way, a read is a plain request for the object.
	before := s.singles.Load()
	if _, err := Get[named](context.Background(), c, "dcim/sites/", 7); err != nil {
		t.Fatal(err)
	}
	if s.singles.Load() != before+1 {
		t.Errorf("a read on its own was not sent as a request for the object")
	}
}

func TestBatchingCanBeTurnedOff(t *testing.T) {
	s := newBatchServer(t)
	c := s.client(t, false)
	for id, err := range readAll(t, c, ids(1, 20)) {
		if err != nil {
			t.Errorf("object %d: %v", id, err)
		}
	}
	if s.lists.Load() != 0 || s.singles.Load() != 20 {
		t.Errorf("got %d single requests and %d lists, want 20 and 0", s.singles.Load(), s.lists.Load())
	}
}

// An object that is not in the list is read alone, so that the caller gets the real 404.
func TestObjectMissingFromABatchIsNotFound(t *testing.T) {
	s := newBatchServer(t)
	s.gone[5], s.gone[6] = true, true
	c := s.client(t, true)
	for id, err := range readAll(t, c, ids(1, 20)) {
		if id == 5 || id == 6 {
			if !IsNotFound(err) {
				t.Errorf("object %d: got %v, want not found", id, err)
			}
		} else if err != nil {
			t.Errorf("object %d: %v", id, err)
		}
	}
	// The list still works for the collection.
	before := s.lists.Load()
	readAll(t, c, ids(21, 40))
	if s.lists.Load() == before {
		t.Errorf("no list request after objects were missing from a batch")
	}
}

// When the list request fails, or does not return what was asked for, every object is read alone and
// the collection is not batched again.
func TestCollectionThatCannotBeListedIsReadObjectByObject(t *testing.T) {
	for name, breakIt := range map[string]func(s *batchServer){
		"list fails":         func(s *batchServer) { s.listStatus = http.StatusForbidden },
		"filter is ignored":  func(s *batchServer) { s.ignoreFilter = true },
		"list returns no ID": func(s *batchServer) { s.ignoreFilter = true; s.gone[1001] = true },
	} {
		t.Run(name, func(t *testing.T) {
			s := newBatchServer(t)
			breakIt(s)
			c := s.client(t, true)
			for id, err := range readAll(t, c, ids(1, 20)) {
				if err != nil {
					t.Errorf("object %d: %v", id, err)
				}
			}
			lists := s.lists.Load()
			if lists == 0 {
				t.Fatalf("no list request was tried")
			}
			for id, err := range readAll(t, c, ids(21, 40)) {
				if err != nil {
					t.Errorf("object %d: %v", id, err)
				}
			}
			if s.lists.Load() != lists {
				t.Errorf("the collection was listed again after a batch did not work")
			}
		})
	}
}

func TestReadsWithDifferentQueriesAreNotMixed(t *testing.T) {
	var mu sync.Mutex
	bad := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		omit := r.URL.Query().Get("omit")
		if strings.HasSuffix(r.URL.Path, "/sites/") {
			var objs []string
			for _, id := range r.URL.Query()["id"] {
				objs = append(objs, fmt.Sprintf(`{"id": %s, "name": %q}`, id, omit))
			}
			_, _ = fmt.Fprintf(w, `{"count": %d, "results": [%s]}`, len(objs), strings.Join(objs, ","))
			return
		}
		id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/dcim/sites"), "/")
		_, _ = fmt.Fprintf(w, `{"id": %s, "name": %q}`, id, omit)
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "token", nil)
	c.SetBatchReads(true)

	var wg sync.WaitGroup
	for i := int64(1); i <= 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			want := ""
			if i%2 == 0 {
				want = "config_context"
			}
			var obj *named
			var err error
			if want == "" {
				obj, err = Get[named](context.Background(), c, "dcim/sites/", i)
			} else {
				obj, err = Get[named](context.Background(), c, "dcim/sites/", i, map[string][]string{"omit": {want}})
			}
			if err != nil || obj.Name != want || obj.ID != i {
				mu.Lock()
				bad++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if bad != 0 {
		t.Errorf("%d reads got the answer to another query", bad)
	}
}

func TestWaitingReadStopsWhenItsContextEnds(t *testing.T) {
	s := newBatchServer(t)
	s.wait = 200 * time.Millisecond
	c := s.client(t, true)

	go func() { _, _ = Get[named](context.Background(), c, "dcim/sites/", 1) }()
	time.Sleep(20 * time.Millisecond) // the first read is under way now

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := Get[named](ctx, c, "dcim/sites/", 2)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want the deadline error", err)
	}
	if time.Since(start) > 150*time.Millisecond {
		t.Errorf("the read waited %s for the batch after its context ended", time.Since(start))
	}
	// Later reads of the collection are not held up by the one that was given up.
	if _, err := Get[named](context.Background(), c, "dcim/sites/", 3); err != nil {
		t.Fatal(err)
	}
}

func TestGetReportsBrokenJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id": "not a number"}`))
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "token", nil)
	if _, err := Get[named](context.Background(), c, "dcim/sites/", 1); err == nil || !strings.Contains(err.Error(), "decoding") {
		t.Fatalf("got %v, want a decoding error", err)
	}
}

// Batches must not wait for each other. When NetBox is slow, holding reads back until the request
// before them is done makes a refresh slower than it is without batching.
func TestBatchesDoNotWaitForEachOther(t *testing.T) {
	s := newBatchServer(t)
	s.wait = 150 * time.Millisecond
	c := s.client(t, true)

	start := time.Now()
	var wg sync.WaitGroup
	// Three groups of reads arrive while the first request is still under way.
	for wave := int64(0); wave < 3; wave++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(wave) * 4 * gatherWindow)
			for id, err := range readAll(t, c, ids(wave*10+1, wave*10+5)) {
				if err != nil {
					t.Errorf("object %d: %v", id, err)
				}
			}
		}()
	}
	wg.Wait()
	// One after the other, the three requests would take 450 ms.
	if took := time.Since(start); took > 300*time.Millisecond {
		t.Errorf("the reads took %s, the batches were sent one after the other", took)
	}
	if n := s.lists.Load(); n != 3 {
		t.Errorf("got %d list requests, want one for each group of reads", n)
	}
}

// A read is held for the gather window at most, however long NetBox takes.
func TestSingleReadIsNotHeldBack(t *testing.T) {
	s := newBatchServer(t)
	s.wait = 0
	c := s.client(t, true)
	start := time.Now()
	for i := int64(1); i <= 5; i++ {
		if _, err := Get[named](context.Background(), c, "dcim/sites/", i); err != nil {
			t.Fatal(err)
		}
	}
	if per := time.Since(start) / 5; per > gatherWindow+20*time.Millisecond {
		t.Errorf("a read on its own took %s", per)
	}
	if s.lists.Load() != 0 {
		t.Errorf("reads on their own were sent as list requests")
	}
}

func TestMoreReadsThanABatchHolds(t *testing.T) {
	// The reads must all arrive inside the gather window, however slowly the goroutines start.
	// With a short window a busy machine splits them into more batches than the test expects.
	window := gatherWindow
	gatherWindow = 200 * time.Millisecond
	t.Cleanup(func() { gatherWindow = window })
	s := newBatchServer(t)
	c := s.client(t, true)
	for id, err := range readAll(t, c, ids(1, 100)) {
		if err != nil {
			t.Errorf("object %d: %v", id, err)
		}
	}
	// Two batches of 50 leave at once. A few reads can be left over for a third and fourth.
	if n := s.lists.Load() + s.singles.Load(); n > 3 {
		t.Errorf("100 reads took %d requests", n)
	}
}

// The reads that do not fit in a full batch wait for more to join, and go out when the window ends.
func TestLeftoverReadsGoOutAfterTheWindow(t *testing.T) {
	window := gatherWindow
	gatherWindow = 100 * time.Millisecond
	t.Cleanup(func() { gatherWindow = window })
	s := newBatchServer(t)
	c := s.client(t, true)
	for id, err := range readAll(t, c, ids(1, 60)) {
		if err != nil {
			t.Errorf("object %d: %v", id, err)
		}
	}
	if n := s.lists.Load() + s.singles.Load(); n > 3 {
		t.Errorf("60 reads took %d requests", n)
	}
}
