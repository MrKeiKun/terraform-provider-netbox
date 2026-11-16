package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// filterBlocks builds the filter blocks of netbox_tenants from name and value pairs.
func filterBlocks(t *testing.T, h *harness, pairs ...string) tftypes.Value {
	filterType := h.schemas.DataSourceSchemas["netbox_tenants"].ValueType().(tftypes.Object).AttributeTypes["filter"]
	elem := filterType.(tftypes.Set).ElementType
	var items []tftypes.Value
	for i := 0; i+1 < len(pairs); i += 2 {
		items = append(items, tftypes.NewValue(elem, map[string]tftypes.Value{"name": str(pairs[i]), "value": str(pairs[i+1])}))
	}
	return tftypes.NewValue(filterType, items)
}

// tenantsQuery reads netbox_tenants with the filters against a NetBox of the given version and
// returns the query of the list request and the diagnostics.
func tenantsQuery(t *testing.T, version string, pairs ...string) (url.Values, []*tfprotov6.Diagnostic) {
	var mu sync.Mutex
	var query url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status/") {
			_, _ = fmt.Fprintf(w, `{"netbox-version": %q}`, version)
			return
		}
		mu.Lock()
		query = r.URL.Query()
		mu.Unlock()
		_, _ = w.Write([]byte(`{"count": 0, "next": null, "results": []}`))
	}))
	defer srv.Close()
	h := newHarness(t, srv.URL)
	typ := h.schemas.DataSourceSchemas["netbox_tenants"].ValueType()
	cfg := h.object(typ, map[string]tftypes.Value{"filter": filterBlocks(t, h, pairs...)})
	resp, err := h.server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "netbox_tenants", Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	return query, resp.Diagnostics
}

// A filter with a lookup expression is sent as the query parameter NetBox knows.
func TestLookupFiltersAreSent(t *testing.T) {
	query, diags := tenantsQuery(t, "4.7.5", "name__ic", "acme", "tag__n", "old,older", "id", "3,4", "id__gt", "2", "tag__any", "a,b")
	for _, d := range diags {
		t.Fatalf("%s: %s", d.Summary, d.Detail)
	}
	for param, want := range map[string]string{
		"name__ic": "acme",
		"tag__n":   "old|older",
		"id":       "3|4",
		"id__gt":   "2",
		"tag__any": "a|b",
	} {
		if got := strings.Join(query[param], "|"); got != want {
			t.Errorf("%s = %q, want %q", param, got, want)
		}
	}
}

// NetBox ignores a filter it does not know and returns every object, so the provider stops before
// the request.
func TestLookupFiltersThatNetBoxDoesNotHaveAreRefused(t *testing.T) {
	for _, tc := range []struct{ version, filter, want string }{
		{"4.7.5", "name__bogus", "has the lookup expressions"},
		{"4.7.5", "nonsense__ic", "not a supported filter"},
		// the tag lookup any came with NetBox 4.6.6
		{"4.6.5", "tag__any", "needs NetBox 4.6.6 or later"},
		{"4.4.0", "tag__any", "needs NetBox 4.6.6 or later"},
	} {
		query, diags := tenantsQuery(t, tc.version, tc.filter, "x")
		found := false
		for _, d := range diags {
			if strings.Contains(d.Detail, tc.want) {
				found = true
			}
		}
		if !found {
			t.Errorf("%s on NetBox %s: no error with %q: %+v", tc.filter, tc.version, tc.want, diags)
		}
		if query != nil {
			t.Errorf("%s on NetBox %s: the request was sent anyway", tc.filter, tc.version)
		}
	}
	if _, diags := tenantsQuery(t, "4.6.6", "tag__any", "x"); len(diags) != 0 {
		t.Errorf("tag__any on NetBox 4.6.6: %+v", diags)
	}
}

func TestWithLookups(t *testing.T) {
	// a data source without filters of its own still gets the id filter and the lookups
	spec := withLookups(&fwresource.DataSourceSpec{Path: "tenancy/tenants/", Plural: &fwresource.Plural{}})
	if spec.Plural.Filters["id"] != "id[]" {
		t.Errorf("no id filter: %v", spec.Plural.Filters)
	}
	if len(spec.Plural.Lookups["name"]) == 0 {
		t.Errorf("no lookups for name")
	}
	// one that declares id itself keeps it
	own := withLookups(&fwresource.DataSourceSpec{Path: "tenancy/tenants/", Plural: &fwresource.Plural{Filters: map[string]string{"id": "id"}}})
	if own.Plural.Filters["id"] != "id" {
		t.Errorf("the id filter of the data source was replaced")
	}
	// singular data sources, ones without filter blocks and unknown paths are left alone
	for _, s := range []*fwresource.DataSourceSpec{
		{Path: "tenancy/tenants/"},
		{Path: "tenancy/tenants/", Plural: &fwresource.Plural{NoFilter: true}},
		{Path: "no/such/path/", Plural: &fwresource.Plural{}},
	} {
		if got := withLookups(s); got.Plural != nil && (got.Plural.Lookups != nil || got.Plural.Filters != nil) {
			t.Errorf("%+v was changed", s)
		}
	}
}
