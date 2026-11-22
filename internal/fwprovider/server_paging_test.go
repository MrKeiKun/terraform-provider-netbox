package fwprovider

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// pagedTenants is a NetBox with count tenants that returns at most maxPage of them for a request,
// as a NetBox does whose administrator lowered MAX_PAGE_SIZE.
func pagedTenants(t *testing.T, count, maxPage int, inflight, peak *atomic.Int32) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status/") {
			_, _ = w.Write([]byte(`{"netbox-version": "4.7.5"}`))
			return
		}
		n := inflight.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		defer inflight.Add(-1)
		time.Sleep(5 * time.Millisecond)

		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if limit <= 0 || limit > maxPage {
			limit = maxPage
		}
		var objs []string
		for i := offset; i < offset+limit && i < count; i++ {
			objs = append(objs, fmt.Sprintf(`{"id": %d, "name": "tenant %d", "slug": "tenant-%d"}`, i+1, i+1, i+1))
		}
		next := "null"
		if offset+limit < count {
			next = `"http://elsewhere.example/api/tenancy/tenants/?limit=` + strconv.Itoa(limit) + `&offset=` + strconv.Itoa(offset+limit) + `"`
		}
		_, _ = fmt.Fprintf(w, `{"count": %d, "next": %s, "results": [%s]}`, count, next, strings.Join(objs, ","))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// readTenantIDs reads netbox_tenants and returns the IDs in the order of the result.
func readTenantIDs(t *testing.T, h *harness, set map[string]tftypes.Value) []int64 {
	typ := h.schemas.DataSourceSchemas["netbox_tenants"].ValueType()
	resp, err := h.server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "netbox_tenants", Config: h.dynamic(h.object(typ, set))})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("read", resp.Diagnostics)
	var list []tftypes.Value
	if err := h.decode(typ, resp.State)["tenants"].As(&list); err != nil {
		t.Fatal(err)
	}
	ids := make([]int64, len(list))
	for i, v := range list {
		var obj map[string]tftypes.Value
		if err := v.As(&obj); err != nil {
			t.Fatal(err)
		}
		var n big.Float
		if err := obj["id"].As(&n); err != nil {
			t.Fatal(err)
		}
		ids[i], _ = n.Int64()
	}
	return ids
}

// A plural data source returns every object once and in the order of NetBox, also when it asks for
// several pages at the same time and NetBox returns fewer objects for a page than asked for.
func TestPluralDataSourceReadsAllPages(t *testing.T) {
	for _, tc := range []struct{ count, maxPage int }{
		{0, 100}, {1, 100}, {100, 100}, {101, 100}, {1037, 100}, {2500, 1000}, {990, 7},
	} {
		t.Run(fmt.Sprintf("%d objects, %d a page", tc.count, tc.maxPage), func(t *testing.T) {
			var inflight, peak atomic.Int32
			srv := pagedTenants(t, tc.count, tc.maxPage, &inflight, &peak)
			h := newHarness(t, srv.URL)
			ids := readTenantIDs(t, h, nil)
			if len(ids) != tc.count {
				t.Fatalf("got %d tenants, want %d", len(ids), tc.count)
			}
			for i, id := range ids {
				if id != int64(i+1) {
					t.Fatalf("result %d has ID %d, want %d: objects are missing, doubled or out of order", i, id, i+1)
				}
			}
			if peak.Load() > 4 {
				t.Errorf("%d pages were asked for at the same time", peak.Load())
			}
			if tc.count > 4*tc.maxPage && peak.Load() < 2 {
				t.Errorf("the pages were asked for one after the other")
			}
		})
	}
}

// limit stops the reading as soon as enough objects are there.
func TestPluralDataSourceLimitStopsPaging(t *testing.T) {
	var inflight, peak atomic.Int32
	srv := pagedTenants(t, 5000, 100, &inflight, &peak)
	var requests atomic.Int32
	counting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		srv.Config.Handler.ServeHTTP(w, r)
	}))
	defer counting.Close()
	h := newHarness(t, counting.URL)
	requests.Store(0)

	ids := readTenantIDs(t, h, map[string]tftypes.Value{"limit": tftypes.NewValue(tftypes.Number, 250)})
	if len(ids) != 250 || ids[0] != 1 || ids[249] != 250 {
		t.Fatalf("got %d tenants, want the first 250", len(ids))
	}
	if n := requests.Load(); n > 3 {
		t.Errorf("reading 250 tenants of 100 a page took %d requests", n)
	}

	// limit = 0 means all, as the documentation says
	if ids := readTenantIDs(t, h, map[string]tftypes.Value{"limit": tftypes.NewValue(tftypes.Number, 0)}); len(ids) != 5000 {
		t.Errorf("limit 0 returned %d tenants, want all 5000", len(ids))
	}
}

// NetBox returns the tagged objects in no fixed order, so the pages of one read can repeat objects and
// leave others out. The data source asks for the order of the IDs, and every object comes back once.
func TestTaggedObjectsAreReadInTheOrderOfTheirIDs(t *testing.T) {
	const count, maxPage = 250, 100
	var orderings atomic.Int32
	var other atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status/") {
			_, _ = w.Write([]byte(`{"netbox-version": "4.7.5"}`))
			return
		}
		if !strings.Contains(r.URL.Path, "tagged-objects") {
			other.Add(1)
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if limit <= 0 || limit > maxPage {
			limit = maxPage
		}
		ordered := r.URL.Query().Get("ordering") == "id"
		if ordered {
			orderings.Add(1)
		}
		var objs []string
		for i := offset; i < offset+limit && i < count; i++ {
			id := i + 1
			if !ordered {
				// without an order, every page starts at another place and some objects are repeated
				id = (i*7+offset)%count + 1
			}
			objs = append(objs, fmt.Sprintf(`{"id": %d, "object": {"id": %d}, "object_id": %d, "object_type": "dcim.site", "tag": {"id": 1}}`, id, id, id))
		}
		next := "null"
		if offset+limit < count {
			next = `"http://elsewhere.example/api/extras/tagged-objects/?limit=` + strconv.Itoa(limit) + `&offset=` + strconv.Itoa(offset+limit) + `"`
		}
		_, _ = fmt.Fprintf(w, `{"count": %d, "next": %s, "results": [%s]}`, count, next, strings.Join(objs, ","))
	}))
	defer srv.Close()
	h := newHarness(t, srv.URL)

	typ := h.schemas.DataSourceSchemas["netbox_tagged_objects"].ValueType()
	resp, err := h.server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "netbox_tagged_objects", Config: h.dynamic(h.object(typ, nil))})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("read", resp.Diagnostics)
	var list []tftypes.Value
	if err := h.decode(typ, resp.State)["tagged_objects"].As(&list); err != nil {
		t.Fatal(err)
	}
	seen := map[int64]bool{}
	for _, v := range list {
		var obj map[string]tftypes.Value
		if err := v.As(&obj); err != nil {
			t.Fatal(err)
		}
		var n big.Float
		if err := obj["id"].As(&n); err != nil {
			t.Fatal(err)
		}
		id, _ := n.Int64()
		if seen[id] {
			t.Fatalf("tagged object %d came back twice", id)
		}
		seen[id] = true
	}
	if len(seen) != count {
		t.Errorf("got %d tagged objects, want %d", len(seen), count)
	}
	if orderings.Load() == 0 {
		t.Error("the tagged objects were not asked for in the order of their IDs")
	}
}

// The endpoints that NetBox leaves unsorted are asked for in the order of their IDs, the others are not.
func TestUnorderedEndpointsAskForTheOrderOfTheirIDs(t *testing.T) {
	for typeName, want := range map[string]bool{
		"netbox_tagged_objects": true,
		"netbox_clusters":       false,
		"netbox_groups":         false,
		"netbox_tenants":        false,
		"netbox_sites":          false,
	} {
		t.Run(typeName, func(t *testing.T) {
			var asked, ordered atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/status/") {
					_, _ = w.Write([]byte(`{"netbox-version": "4.7.5"}`))
					return
				}
				asked.Add(1)
				if r.URL.Query().Get("ordering") == "id" {
					ordered.Add(1)
				}
				_, _ = w.Write([]byte(`{"count": 0, "next": null, "results": []}`))
			}))
			defer srv.Close()
			h := newHarness(t, srv.URL)
			typ := h.schemas.DataSourceSchemas[typeName].ValueType()
			resp, err := h.server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: typeName, Config: h.dynamic(h.object(typ, nil))})
			if err != nil {
				t.Fatal(err)
			}
			h.noErrors("read", resp.Diagnostics)
			if asked.Load() == 0 {
				t.Fatal("NetBox was not asked for the list")
			}
			if got := ordered.Load() > 0; got != want {
				t.Errorf("asked for the order of the IDs = %v, want %v", got, want)
			}
		})
	}
}
