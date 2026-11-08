package netboxapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoRetriesTemporaryFailures(t *testing.T) {
	retryDelay = time.Millisecond
	defer func() { retryDelay = time.Second }()

	tests := []struct {
		name     string
		method   string
		failures int32
		status   int
		wantErr  bool
		wantHits int32
		noRetry  bool
	}{
		{"GET recovers after 503", http.MethodGet, 2, http.StatusServiceUnavailable, false, 3, false},
		{"GET gives up after three tries", http.MethodGet, 5, http.StatusBadGateway, true, 3, false},
		{"POST is never repeated", http.MethodPost, 1, http.StatusServiceUnavailable, true, 1, false},
		{"GET does not retry a 404", http.MethodGet, 1, http.StatusNotFound, true, 1, false},
		{"PUT is repeated", http.MethodPut, 1, http.StatusGatewayTimeout, false, 2, false},
		{"retries can be turned off", http.MethodGet, 1, http.StatusServiceUnavailable, true, 1, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var hits int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if atomic.AddInt32(&hits, 1) <= tc.failures {
					w.WriteHeader(tc.status)
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
			if tc.method == http.MethodPut || tc.method == http.MethodPost {
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
