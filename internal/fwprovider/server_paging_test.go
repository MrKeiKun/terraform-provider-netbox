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
