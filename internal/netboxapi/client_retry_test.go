package netboxapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// deadlockBody is what NetBox returns when Postgres aborts a request because of a deadlock.
const deadlockBody = `{"error": "deadlock detected
DETAIL:  Process 221 waits for ShareLock on transaction 1892; blocked by process 200."}`

func TestDoRetriesTemporaryFailures(t *testing.T) {
	retryDelay = time.Millisecond
	defer func() { retryDelay = time.Second }()

	tests := []struct {
		name     string
		method   string
		failures int32
		status   int
		body     string
		wantErr  bool
		wantHits int32
		noRetry  bool
	}{
		{"GET recovers after 503", http.MethodGet, 2, http.StatusServiceUnavailable, "", false, 3, false},
		{"GET gives up after three tries", http.MethodGet, 5, http.StatusBadGateway, "", true, 3, false},
		{"POST is never repeated", http.MethodPost, 1, http.StatusServiceUnavailable, "", true, 1, false},
		{"GET does not retry a 404", http.MethodGet, 1, http.StatusNotFound, "", true, 1, false},
		{"PUT is repeated", http.MethodPut, 1, http.StatusGatewayTimeout, "", false, 2, false},
		{"retries can be turned off", http.MethodGet, 1, http.StatusServiceUnavailable, "", true, 1, true},
		{"DELETE recovers after a deadlock", http.MethodDelete, 2, http.StatusInternalServerError, deadlockBody, false, 3, false},
		{"PUT recovers after a deadlock", http.MethodPut, 1, http.StatusInternalServerError, deadlockBody, false, 2, false},
		{"POST is repeated after a deadlock", http.MethodPost, 1, http.StatusInternalServerError, deadlockBody, false, 2, false},
		{"PATCH is repeated after a deadlock", http.MethodPatch, 1, http.StatusInternalServerError, deadlockBody, false, 2, false},
		{"POST is not repeated after another 500", http.MethodPost, 1, http.StatusInternalServerError, `{"error": "something else"}`, true, 1, false},
		{"other 500 errors are not retried", http.MethodDelete, 1, http.StatusInternalServerError, `{"error": "something else"}`, true, 1, false},
		{"DELETE gives up after repeated deadlocks", http.MethodDelete, 5, http.StatusInternalServerError, deadlockBody, true, 3, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var hits int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if atomic.AddInt32(&hits, 1) <= tc.failures {
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte(tc.body))
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			c, err := New(srv.URL, "token", nil)
			if err != nil {
				t.Fatal(err)
			}
			c.SetMaxRetries(2)
			if tc.noRetry {
				c.SetMaxRetries(0)
			}
			var body any
			if tc.method == http.MethodPut || tc.method == http.MethodPost || tc.method == http.MethodPatch {
				body = map[string]string{"a": "b"}
			}
			var out map[string]any
			err = c.Do(context.Background(), tc.method, srv.URL+"/api/x/", body, &out)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, want error: %v", err, tc.wantErr)
			}
			if hits != tc.wantHits {
				t.Fatalf("server was hit %d times, want %d", hits, tc.wantHits)
			}
		})
	}
}

func TestDoStopsRetryingWhenContextIsCancelled(t *testing.T) {
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
	c.SetMaxRetries(2)

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	done := make(chan error, 1)
	go func() { done <- c.Do(ctx, http.MethodGet, srv.URL+"/api/x/", nil, nil) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected an error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Do kept waiting after the context was cancelled")
	}
	if hits != 1 {
		t.Fatalf("server was hit %d times, want 1", hits)
	}
}
