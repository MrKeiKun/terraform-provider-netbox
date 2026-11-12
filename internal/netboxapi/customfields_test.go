package netboxapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestTypedCustomFieldValue(t *testing.T) {
	tests := []struct {
		typ, in string
		want    any
	}{
		{"integer", "5", int64(5)},
		{"integer", "abc", "abc"},
		{"integer", "", nil},
		{"boolean", "true", true},
		{"boolean", "false", false},
		{"decimal", "1.50", json.Number("1.50")},
		{"text", "hello", "hello"},
		{"text", "", nil},
		{"select", "a", "a"},
		{"select", `{"value":"a","label":"A"}`, "a"},
		{"multiselect", `["a","b"]`, []any{"a", "b"}},
		{"multiselect", "a, b", []any{"a", "b"}},
		{"multiselect", `[{"value":"a","label":"A"}]`, []any{"a"}},
		{"object", "12", int64(12)},
		{"object", `{"id":12,"url":"u"}`, int64(12)},
		{"multiobject", `[1,2]`, []any{int64(1), int64(2)}},
		{"multiobject", "1,2", []any{int64(1), int64(2)}},
		{"multiobject", `[{"id":3,"url":"u"}]`, []any{int64(3)}},
		{"json", `{"a": 1}`, json.RawMessage(`{"a": 1}`)},
		{"json", `not json`, "not json"},
		{"unknown", "x", "x"},
	}
	for _, tt := range tests {
		t.Run(tt.typ+"/"+tt.in, func(t *testing.T) {
			got := TypedCustomFieldValue(tt.typ, tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestClientTypesCustomFieldsInRequests(t *testing.T) {
	var lookups int32
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/extras/custom-fields/":
			atomic.AddInt32(&lookups, 1)
			_, _ = w.Write([]byte(`{"count":2,"next":null,"results":[
				{"name":"num","type":{"value":"integer"}},
				{"name":"flag","type":{"value":"boolean"}}]}`))
		default:
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	c, err := New(srv.URL, "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	body := map[string]any{"name": "x", "custom_fields": map[string]any{"num": "5", "flag": "true", "other": "s"}}
	for i := 0; i < 2; i++ {
		if _, err := Create[map[string]any](context.Background(), c, "dcim/sites/", body); err != nil {
			t.Fatal(err)
		}
	}
	// "other" has no definition, so it triggers a refresh the first time only
	// per request; the definitions are not cached for unknown names.
	cf := gotBody["custom_fields"].(map[string]any)
	if cf["num"] != float64(5) || cf["flag"] != true || cf["other"] != "s" {
		t.Errorf("custom_fields not typed: %v", cf)
	}
	if atomic.LoadInt32(&lookups) == 0 {
		t.Error("expected the custom field definitions to be fetched")
	}
}

// A list of objects, as sent when NetBox allocates addresses, gets typed custom fields too.
func TestClientTypesCustomFieldsInListBodies(t *testing.T) {
	var got []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"count":1,"next":null,"results":[{"name":"num","type":{"value":"integer"}}]}`))
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	c, err := New(srv.URL, "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	body := []any{map[string]any{"custom_fields": map[string]any{"num": "5"}}}
	if _, err := Request[[]map[string]any](context.Background(), c, http.MethodPost, "ipam/prefixes/1/available-ips/", nil, body); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %v", got)
	}
	if cf := got[0]["custom_fields"].(map[string]any); cf["num"] != float64(5) {
		t.Errorf("custom field was not typed: %v", cf)
	}
}
