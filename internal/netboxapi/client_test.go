package netboxapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type tag struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
}

func newTestClient(t *testing.T, h http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := New(srv.URL, "abc", srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestAuthHeader(t *testing.T) {
	for token, want := range map[string]string{"abc": "Token abc", "nbt_x.y": "Bearer nbt_x.y"} {
		c := &Client{token: token}
		if got := c.authHeader(); got != want {
			t.Errorf("token %q: got %q, want %q", token, got, want)
		}
	}
}

func TestCRUD(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token abc" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/extras/tags/":
			var in tag
			_ = json.NewDecoder(r.Body).Decode(&in)
			in.ID = 7
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(in)
		case r.Method == http.MethodGet && r.URL.Path == "/api/extras/tags/7/":
			_ = json.NewEncoder(w).Encode(tag{ID: 7, Name: "a"})
		case r.Method == http.MethodPut && r.URL.Path == "/api/extras/tags/7/":
			var in tag
			_ = json.NewDecoder(r.Body).Decode(&in)
			in.ID = 7
			_ = json.NewEncoder(w).Encode(in)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/extras/tags/7/":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"Not found."}`))
		}
	}))
	ctx := context.Background()

	created, err := Create[tag](ctx, c, "extras/tags/", tag{Name: "a"})
	if err != nil || created.ID != 7 {
		t.Fatalf("create: %v %+v", err, created)
	}
	got, err := Get[tag](ctx, c, "extras/tags/", 7)
	if err != nil || got.Name != "a" {
		t.Fatalf("get: %v %+v", err, got)
	}
	upd, err := Update[tag](ctx, c, "extras/tags/", 7, tag{Name: "b"})
	if err != nil || upd.Name != "b" {
		t.Fatalf("update: %v %+v", err, upd)
	}
	if err := Delete(ctx, c, "extras/tags/", 7); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := Get[tag](ctx, c, "extras/tags/", 99); !IsNotFound(err) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestListFollowsPagination(t *testing.T) {
	var srvURL string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") == "1" {
			_ = json.NewEncoder(w).Encode(map[string]any{"next": nil, "results": []tag{{ID: 2, Name: "b"}}})
			return
		}
		next := srvURL + "/api/extras/tags/?offset=1"
		_ = json.NewEncoder(w).Encode(map[string]any{"next": next, "results": []tag{{ID: 1, Name: "a"}}})
	}))
	srvURL = c.baseURL.String()

	got, err := List[tag](context.Background(), c, "extras/tags/", nil)
	if err != nil || len(got) != 2 || got[1].Name != "b" {
		t.Fatalf("list: %v %+v", err, got)
	}
}

// The link to the next page may name another host. The request goes to the configured server.
func TestListIgnoresHostOfNextLink(t *testing.T) {
	evil := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the request went to the host named in the response")
	}))
	defer evil.Close()

	var queries []url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		if r.URL.Query().Get("offset") == "" {
			_ = json.NewEncoder(w).Encode(map[string]any{"next": evil.URL + "/api/dcim/sites/?limit=1&offset=1&name=x", "results": []any{map[string]any{"id": 1}}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"next": nil, "results": []any{map[string]any{"id": 2}}})
	}))
	defer srv.Close()

	c, _ := New(srv.URL, "token", nil)
	all, err := List[map[string]any](context.Background(), c, "dcim/sites/", url.Values{"name": {"x"}, "limit": {"1"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("got %d objects, want 2", len(all))
	}
	if len(queries) != 2 || queries[1].Get("offset") != "1" || queries[1].Get("name") != "x" {
		t.Errorf("queries = %v", queries)
	}
}

func TestCRUDErrors(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"no"}`))
	}))
	ctx := context.Background()
	if _, err := Create[tag](ctx, c, "extras/tags/", tag{Name: "a"}); err == nil {
		t.Error("create")
	}
	if _, err := Update[tag](ctx, c, "extras/tags/", 1, tag{Name: "a"}); err == nil {
		t.Error("update")
	}
	if _, err := Patch[tag](ctx, c, "extras/tags/", 1, tag{Name: "a"}); err == nil {
		t.Error("patch")
	}
	if _, err := Get[tag](ctx, c, "extras/tags/", 1); err == nil {
		t.Error("get")
	}
	if _, err := ListPage[tag](ctx, c, "extras/tags/", nil); err == nil {
		t.Error("list page")
	}
	if _, err := List[tag](ctx, c, "extras/tags/", nil); err == nil {
		t.Error("list")
	}
}

func TestListRejectsInvalidNextLink(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"next":"http://bad host/%zz","results":[]}`))
	}))
	if _, err := List[tag](context.Background(), c, "extras/tags/", nil); err == nil {
		t.Error("an invalid next page link was followed")
	}
}
