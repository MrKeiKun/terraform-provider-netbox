package fwprovider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func connectTo(t *testing.T, url string) *providerstate.State {
	t.Helper()
	st, diags := providerstate.Connect(context.Background(), providerstate.Settings{
		ServerURL: url, APIToken: "0123456789abcdef0123456789abcdef01234567", RequestTimeout: 10, SkipVersionCheck: true,
	})
	if st == nil {
		t.Fatalf("could not connect: %v", diags)
	}
	st.Rest.SetMaxRetries(0)
	return st
}

func TestAvailableIPAllocationErrors(t *testing.T) {
	allocate := availableIPAddress().Allocate
	ctx := context.Background()

	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer empty.Close()
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	defer failing.Close()

	plan := fwresource.Values{"prefix_id": types.Int64Value(1)}
	if _, err := allocate(ctx, connectTo(t, empty.URL), plan, map[string]any{"vrf": int64(2), "tenant": int64(3)}); err == nil {
		t.Error("an empty prefix was accepted")
	}
	if _, err := allocate(ctx, connectTo(t, failing.URL), plan, map[string]any{}); err == nil {
		t.Error("a failing NetBox was accepted")
	}
	if _, err := allocate(ctx, connectTo(t, empty.URL), fwresource.Values{}, map[string]any{}); err == nil {
		t.Error("an allocation without a prefix or range was accepted")
	}
	if got := stringValue(fwresource.Values{"s": types.Int64Value(1)}, "s"); got != "" {
		t.Errorf("stringValue of a number = %q", got)
	}
}

func TestFindParentPrefixID(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	fake.put("ipam/prefixes", map[string]any{"id": float64(1), "prefix": "10.0.0.0/16"})
	fake.put("ipam/prefixes", map[string]any{"id": float64(2), "prefix": "10.0.0.0/24"})
	fake.put("ipam/prefixes", map[string]any{"id": float64(3), "prefix": "not a prefix"})
	st := connectTo(t, srv.URL)
	ctx := context.Background()

	if id, ok := findParentPrefixID(ctx, st, map[string]any{"address": "10.0.0.5/24"}); !ok || id != 2 {
		t.Errorf("most specific prefix = %d, %v", id, ok)
	}
	if id, ok := findParentPrefixID(ctx, st, map[string]any{"address": "10.0.0.5/24", "vrf": map[string]any{"id": float64(7)}}); !ok || id != 2 {
		t.Errorf("prefix in a VRF = %d, %v", id, ok)
	}
	if _, ok := findParentPrefixID(ctx, st, map[string]any{"address": "nonsense"}); ok {
		t.Error("an address that is not valid found a prefix")
	}
	fake.failOn("GET", "ipam/prefixes", 500)
	if _, ok := findParentPrefixID(ctx, st, map[string]any{"address": "10.0.0.5/24"}); ok {
		t.Error("a failing NetBox found a prefix")
	}
}
