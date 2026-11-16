package fwprovider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// refreshAll creates count tenants and then refreshes them side by side, as Terraform does for a
// plan. It returns the number of GET requests the refresh took.
func refreshAll(t *testing.T, count int) int {
	fake, inner := newFakeNetBox()
	defer inner.Close()
	// NetBox takes a moment to answer, so reads pile up while one is under way.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			time.Sleep(5 * time.Millisecond)
		}
		fake.ServeHTTP(w, r)
	}))
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_tenant"

	states := make([]tftypes.Value, count)
	for i := range states {
		states[i] = h.apply(name, tftypes.NewValue(tftypes.Object{}, nil), map[string]tftypes.Value{
			"name":        str(fmt.Sprintf("tenant %d", i)),
			"description": str(fmt.Sprintf("description %d", i)),
		})
	}

	fake.mu.Lock()
	fake.calls = nil
	fake.mu.Unlock()

	responses := make([]*tfprotov6.ReadResourceResponse, count)
	var wg sync.WaitGroup
	for i := range states {
		wg.Add(1)
		go func() {
			defer wg.Done()
			responses[i] = h.read(name, states[i])
		}()
	}
	wg.Wait()

	typ := h.schemas.ResourceSchemas[name].ValueType()
	for i, rr := range responses {
		h.noErrors("read", rr.Diagnostics)
		back := h.decode(typ, rr.NewState)
		for _, attr := range []string{"id", "name", "description", "slug"} {
			if back[attr].String() != attrOf(t, states[i], attr).String() {
				t.Errorf("tenant %d: %s changed on refresh: %s", i, attr, back[attr])
			}
		}
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	gets := 0
	for _, call := range fake.calls {
		if strings.HasPrefix(call, http.MethodGet+" ") {
			gets++
		}
	}
	return gets
}

// The provider reads resources that are refreshed at the same time with fewer requests than
// resources, and every resource still gets its own object.
func TestRefreshBatchesReads(t *testing.T) {
	const count = 30
	if gets := refreshAll(t, count); gets > count/2 {
		t.Errorf("refreshing %d resources took %d requests", count, gets)
	}
}

func TestBatchReadsCanBeTurnedOff(t *testing.T) {
	t.Setenv("NETBOX_BATCH_READS", "false")
	const count = 30
	if gets := refreshAll(t, count); gets != count {
		t.Errorf("refreshing %d resources took %d requests, want one for each", count, gets)
	}
}

// When the answer to a create is lost, the object may exist in NetBox without Terraform knowing it.
// The error has to say so, or the next apply creates a second object.
func TestCreateErrorTellsWhenTheObjectMayExist(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_contact_role"
	cfg := map[string]tftypes.Value{"name": str("x")}
	const hint = "may have created the object"

	for status, want := range map[int]bool{
		http.StatusGatewayTimeout: true,
		http.StatusBadGateway:     true,
		http.StatusBadRequest:     false,
		http.StatusForbidden:      false,
	} {
		fake.failOn("POST", "tenancy/contact-roles", status)
		diags, _ := h.tryApply(name, nullObject(), cfg)
		h.wantError(fmt.Sprintf("a create answered with %d", status), diags)
		got := false
		for _, d := range diags {
			if strings.Contains(d.Detail, hint) {
				got = true
			}
		}
		if got != want {
			t.Errorf("status %d: hint in the error = %v, want %v", status, got, want)
		}
		fake.clearFailures()
	}
}
