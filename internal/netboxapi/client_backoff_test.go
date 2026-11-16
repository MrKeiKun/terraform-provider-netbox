package netboxapi

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackoffGrowsAndStaysInBounds(t *testing.T) {
	retryDelay = time.Second
	for attempt, base := range map[int]time.Duration{1: time.Second, 2: 2 * time.Second, 3: 4 * time.Second, 4: 8 * time.Second, 20: maxRetryDelay} {
		for i := 0; i < 200; i++ {
			d := backoff(attempt, nil)
			if d < base-base/4 || d > base+base/4 {
				t.Fatalf("attempt %d: wait %s is not within a quarter of %s", attempt, d, base)
			}
		}
	}
}

func TestBackoffUsesRetryAfter(t *testing.T) {
	if d := backoff(1, &APIError{RetryAfter: 7 * time.Second}); d != 7*time.Second {
		t.Fatalf("got %s, want the wait the server asked for", d)
	}
	if d := backoff(1, &APIError{RetryAfter: time.Hour}); d != maxRetryDelay {
		t.Fatalf("got %s, want the longest wait %s", d, maxRetryDelay)
	}
}

func TestRetryAfterHeader(t *testing.T) {
	h := http.Header{}
	if d := retryAfter(h); d != 0 {
		t.Fatalf("no header: got %s", d)
	}
	h.Set("Retry-After", "3")
	if d := retryAfter(h); d != 3*time.Second {
		t.Fatalf("seconds: got %s", d)
	}
	h.Set("Retry-After", "-3")
	if d := retryAfter(h); d != 0 {
		t.Fatalf("negative seconds: got %s", d)
	}
	h.Set("Retry-After", time.Now().Add(time.Minute).UTC().Format(http.TimeFormat))
	if d := retryAfter(h); d < 50*time.Second || d > time.Minute {
		t.Fatalf("date: got %s", d)
	}
	h.Set("Retry-After", "soon")
	if d := retryAfter(h); d != 0 {
		t.Fatalf("garbage: got %s", d)
	}
}

// 429 means the request was turned away before NetBox worked on it, so every method is sent again.
func TestTooManyRequestsIsRetriedForEveryMethod(t *testing.T) {
	retryDelay = time.Millisecond
	defer func() { retryDelay = time.Second }()

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete} {
		var hits int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if atomic.AddInt32(&hits, 1) == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			_, _ = w.Write([]byte(`{}`))
		}))
		c, err := New(srv.URL, "token", nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := c.Do(context.Background(), method, srv.URL+"/api/x/", nil, nil); err != nil {
			t.Errorf("%s: %v", method, err)
		}
		if hits != 2 {
			t.Errorf("%s: %d requests, want 2", method, hits)
		}
		srv.Close()
	}
}

func TestUserAgentAndQueryAreSent(t *testing.T) {
	var gotUA, gotQuery, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA, gotQuery, gotMethod = r.Header.Get("User-Agent"), r.URL.RawQuery, r.Method
		_, _ = w.Write([]byte(`{"id": 1}`))
	}))
	defer srv.Close()

	c, err := New(srv.URL, "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	c.SetUserAgent("terraform-provider-netbox/test")
	ctx := context.Background()
	q := url.Values{"omit": {"config_context"}}

	calls := map[string]func() error{
		http.MethodGet:   func() error { _, err := Get[Ref](ctx, c, "dcim/devices/", 1, q); return err },
		http.MethodPost:  func() error { _, err := Create[Ref](ctx, c, "dcim/devices/", map[string]any{}, q); return err },
		http.MethodPut:   func() error { _, err := Update[Ref](ctx, c, "dcim/devices/", 1, map[string]any{}, q); return err },
		http.MethodPatch: func() error { _, err := Patch[Ref](ctx, c, "dcim/devices/", 1, map[string]any{}, q); return err },
	}
	for method, call := range calls {
		if err := call(); err != nil {
			t.Fatal(err)
		}
		if gotMethod != method || gotQuery != "omit=config_context" {
			t.Errorf("%s: got %s with query %q", method, gotMethod, gotQuery)
		}
		if gotUA != "terraform-provider-netbox/test" {
			t.Errorf("%s: User-Agent %q", method, gotUA)
		}
	}
	if _, err := Get[Ref](ctx, c, "dcim/devices/", 1); err != nil || gotQuery != "" {
		t.Errorf("without query: err %v, query %q", err, gotQuery)
	}
}

func TestLongErrorBodyIsCut(t *testing.T) {
	err := &APIError{Method: "GET", URL: "http://x/api/a/", StatusCode: 502, Body: strings.Repeat("é", 5000)}
	msg := err.Error()
	if len(msg) > maxErrorBody+200 {
		t.Fatalf("message has %d bytes", len(msg))
	}
	if !strings.Contains(msg, "more bytes)") || strings.ContainsRune(msg, '�') {
		t.Fatalf("message is not cut cleanly: %q", msg[len(msg)-40:])
	}
	short := &APIError{Method: "GET", URL: "u", StatusCode: 400, Body: "bad"}
	if got := short.Error(); got != "netbox: GET u returned 400: bad" {
		t.Fatalf("got %q", got)
	}
}

// A body without custom fields goes out as it was built, and does not need the field definitions.
func TestBodyWithoutCustomFieldsIsNotConverted(t *testing.T) {
	var lists int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "custom-fields") {
			atomic.AddInt32(&lists, 1)
		}
		_, _ = w.Write([]byte(`{"id": 1}`))
	}))
	defer srv.Close()
	c, err := New(srv.URL, "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Create[Ref](context.Background(), c, "dcim/sites/", map[string]any{"name": "a"}); err != nil {
		t.Fatal(err)
	}
	if lists != 0 {
		t.Fatalf("the custom field definitions were read %d times", lists)
	}
}

func BenchmarkEncodeBody(b *testing.B) {
	c, _ := New("http://netbox.example", "token", nil)
	body := map[string]any{"name": "site", "slug": "site", "status": "active", "description": strings.Repeat("x", 200), "tags": []TagRef{{Name: "a", Slug: "a"}, {Name: "b", Slug: "b"}}}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := c.encodeBody(context.Background(), http.MethodPost, "http://netbox.example/api/dcim/sites/", body); err != nil {
			b.Fatal(err)
		}
	}
}

// When no connection can be made nothing was sent, so a request that creates an object is sent again
// too. NetBox is briefly unreachable when it restarts.
func TestRequestIsRepeatedWhenNoConnectionCanBeMade(t *testing.T) {
	retryDelay = time.Millisecond
	defer func() { retryDelay = time.Second }()

	// A port nothing listens on.
	l, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()

	var dials atomic.Int32
	hc := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			dials.Add(1)
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	}}
	c, err := New("http://"+addr, "token", hc)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Do(context.Background(), http.MethodPost, "http://"+addr+"/api/dcim/sites/", map[string]any{"name": "a"}, nil)
	if err == nil {
		t.Fatal("no error for a server that is not there")
	}
	if got := dials.Load(); got != 1+DefaultMaxRetries {
		t.Errorf("%d connection attempts, want %d", got, 1+DefaultMaxRetries)
	}
}

// A connection that breaks after the request went out may have reached NetBox, so a POST is not
// sent again.
func TestPostIsNotRepeatedWhenTheConnectionBreaks(t *testing.T) {
	retryDelay = time.Millisecond
	defer func() { retryDelay = time.Second }()

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		conn, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			_ = conn.Close()
		}
	}))
	defer srv.Close()
	c, err := New(srv.URL, "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Do(context.Background(), http.MethodPost, srv.URL+"/api/dcim/sites/", map[string]any{"name": "a"}, nil); err == nil {
		t.Fatal("no error for a broken connection")
	}
	if got := hits.Load(); got != 1 {
		t.Errorf("the POST was sent %d times", got)
	}
}

func TestMaybeProcessed(t *testing.T) {
	for name, tc := range map[string]struct {
		err  error
		want bool
	}{
		"no error":          {nil, false},
		"validation error":  {&APIError{StatusCode: http.StatusBadRequest}, false},
		"too many requests": {&APIError{StatusCode: http.StatusTooManyRequests}, false},
		"unavailable":       {&APIError{StatusCode: http.StatusServiceUnavailable}, false},
		"bad gateway":       {&APIError{StatusCode: http.StatusBadGateway}, true},
		"gateway timeout":   {&APIError{StatusCode: http.StatusGatewayTimeout}, true},
		"no connection":     {&net.OpError{Op: "dial"}, false},
		"broken connection": {&net.OpError{Op: "read"}, true},
		"timeout":           {context.DeadlineExceeded, true},
	} {
		if got := MaybeProcessed(tc.err); got != tc.want {
			t.Errorf("%s: got %v, want %v", name, got, tc.want)
		}
	}
}
