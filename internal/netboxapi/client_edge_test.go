package netboxapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewRejectsBadInput(t *testing.T) {
	if _, err := New("http://localhost", "", nil); err == nil {
		t.Error("expected an error for an empty token")
	}
	if _, err := New("http://[::1", "token", nil); err == nil {
		t.Error("expected an error for an unparsable URL")
	}
}

func TestSetMaxRetriesIgnoresNegativeValues(t *testing.T) {
	c, err := New("http://localhost", "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	c.SetMaxRetries(-5)
	if c.maxRetries != 0 {
		t.Fatalf("maxRetries = %d, want 0", c.maxRetries)
	}
}

func TestDoRepeatsTheRequestBody(t *testing.T) {
	retryDelay = time.Millisecond
	defer func() { retryDelay = time.Second }()

	var hits int32
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if atomic.AddInt32(&hits, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := New(srv.URL, "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Do(context.Background(), http.MethodPut, srv.URL+"/api/x/", map[string]string{"a": "b"}, nil); err != nil {
		t.Fatal(err)
	}
	if len(bodies) != 2 || bodies[0] != bodies[1] || bodies[0] == "" {
		t.Fatalf("request bodies = %q, want the same non-empty body twice", bodies)
	}
}

func TestDoRetriesConnectionErrors(t *testing.T) {
	retryDelay = time.Millisecond
	defer func() { retryDelay = time.Second }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close() // nothing listens any more

	c, err := New(url, "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	c.SetMaxRetries(1)
	if err := c.Do(context.Background(), http.MethodGet, url+"/api/x/", nil, nil); err == nil {
		t.Fatal("expected an error for an unreachable server")
	}
}

func TestDoStopsWaitingWhenContextIsCancelled(t *testing.T) {
	retryDelay = time.Hour
	defer func() { retryDelay = time.Second }()

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c, err := New(srv.URL, "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := c.Do(ctx, http.MethodGet, srv.URL+"/api/x/", nil, nil); err == nil {
		t.Fatal("expected an error")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("Do kept waiting after the context was cancelled")
	}
	if hits != 1 {
		t.Fatalf("server was hit %d times, want 1", hits)
	}
}

func TestDoReportsUndecodableResponses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	c, err := New(srv.URL, "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := c.Do(context.Background(), http.MethodGet, srv.URL+"/api/x/", nil, &out); err == nil {
		t.Fatal("expected a decoding error")
	}
}

func TestDoReportsUnencodableBodies(t *testing.T) {
	c, err := New("http://localhost", "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Do(context.Background(), http.MethodPost, "http://localhost/api/x/", make(chan int), nil); err == nil {
		t.Fatal("expected an encoding error")
	}
}
