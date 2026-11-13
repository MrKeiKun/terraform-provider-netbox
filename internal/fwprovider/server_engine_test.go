package fwprovider

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func setOf(elem tftypes.Type, vals ...tftypes.Value) tftypes.Value {
	return tftypes.NewValue(tftypes.Set{ElementType: elem}, vals)
}

func TestTagsAndCustomFields(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_tenant"

	fake.put("extras/tags", map[string]any{"id": float64(1), "name": "blue", "slug": "blue"})

	state := h.apply(name, nullObject(), map[string]tftypes.Value{
		"name": str("tagged"),
		"tags": setOf(tftypes.String, str("blue")),
		"custom_fields": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
			"owner": str("ops"),
		}),
	})
	obj := fake.get("tenancy/tenants", int(numberAttr(t, state, "id")))
	if tags, _ := obj["tags"].([]any); len(tags) != 1 {
		t.Errorf("tags sent to NetBox = %v", obj["tags"])
	}
	if cf, _ := obj["custom_fields"].(map[string]any); cf["owner"] != "ops" {
		t.Errorf("custom fields sent to NetBox = %v", obj["custom_fields"])
	}

	// the tags come back from NetBox as objects
	obj["tags"] = []any{map[string]any{"id": float64(1), "name": "blue", "slug": "blue"}}
	rr := h.read(name, state)
	h.noErrors("read", rr.Diagnostics)
	got := h.decode(h.schemas.ResourceSchemas[name].ValueType(), rr.NewState)
	var tags []tftypes.Value
	if err := got["tags"].As(&tags); err != nil || len(tags) != 1 {
		t.Errorf("tags in the state = %v, %v", got["tags"], err)
	}

	// removing the tags clears them
	h.apply(name, state, map[string]tftypes.Value{"name": str("tagged")})
	if tags, _ := obj["tags"].([]any); len(tags) != 0 {
		t.Errorf("tags after removal = %v", obj["tags"])
	}
}

func numberAttr(t *testing.T, state tftypes.Value, name string) int64 {
	t.Helper()
	s := stringAttr(t, state, name)
	var n int64
	for _, c := range s {
		n = n*10 + int64(c-'0')
	}
	return n
}

func TestDefaultTags(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	fake.put("extras/tags", map[string]any{"id": float64(1), "name": "managed", "slug": "managed"})

	server := newHarness(t, srv.URL)
	// configure again with default tags
	cfg := server.object(server.schemas.Provider.ValueType(), map[string]tftypes.Value{
		"server_url":   str(srv.URL),
		"api_token":    str("token"),
		"default_tags": setOf(tftypes.String, str("managed")),
	})
	resp, err := server.server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: server.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	server.noErrors("configure", resp.Diagnostics)

	state := server.apply("netbox_tenant", nullObject(), map[string]tftypes.Value{"name": str("defaults")})
	obj := fake.get("tenancy/tenants", int(numberAttr(t, state, "id")))
	if tags, _ := obj["tags"].([]any); len(tags) != 1 {
		t.Errorf("default tag was not sent: %v", obj["tags"])
	}
}

func TestScopedResource(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_cluster"

	fake.put("virtualization/cluster-types", map[string]any{"id": float64(1), "name": "kvm"})
	fake.put("dcim/sites", map[string]any{"id": float64(2), "name": "s"})

	state := h.apply(name, nullObject(), map[string]tftypes.Value{
		"name":            str("c1"),
		"cluster_type_id": num(1),
		"site_id":         num(2),
	})
	obj := fake.get("virtualization/clusters", int(numberAttr(t, state, "id")))
	if obj["scope_type"] != "dcim.site" || obj["scope_id"] != float64(2) {
		t.Errorf("scope sent to NetBox: type %v, id %v", obj["scope_type"], obj["scope_id"])
	}

	// the scope comes back as the attribute of its type
	obj["scope_type"] = "dcim.site"
	obj["scope_id"] = float64(2)
	rr := h.read(name, state)
	h.noErrors("read", rr.Diagnostics)
	got := h.decode(h.schemas.ResourceSchemas[name].ValueType(), rr.NewState)
	var site big.Float
	if err := got["site_id"].As(&site); err != nil {
		t.Fatalf("site_id: %v", err)
	}
	if n, _ := site.Int64(); n != 2 {
		t.Errorf("site_id = %s, want 2", site.String())
	}

	// moving to another scope type
	fake.put("dcim/locations", map[string]any{"id": float64(3), "name": "l"})
	state = h.apply(name, state, map[string]tftypes.Value{
		"name":            str("c1"),
		"cluster_type_id": num(1),
		"location_id":     num(3),
	})
	if obj["scope_type"] != "dcim.location" {
		t.Errorf("scope type after the change = %v", obj["scope_type"])
	}
	_ = state
}

func TestAvailableIPAddress(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_available_ip_address"

	fake.put("ipam/prefixes", map[string]any{"id": float64(1), "prefix": "10.0.0.0/24"})

	state := h.apply(name, nullObject(), map[string]tftypes.Value{"prefix_id": num(1), "status": str("reserved")})
	ip := stringAttr(t, state, "ip_address")
	if ip == "" {
		t.Fatal("no address was allocated")
	}
	obj := fake.get("ipam/ip-addresses", int(numberAttr(t, state, "id")))
	if obj["status"] != "reserved" {
		t.Errorf("status = %v", obj["status"])
	}

	// the address is kept when the object is updated
	state = h.apply(name, state, map[string]tftypes.Value{"prefix_id": num(1), "status": str("active")})
	if got := stringAttr(t, state, "ip_address"); got != ip {
		t.Errorf("address changed from %s to %s", ip, got)
	}
	h.noErrors("destroy", h.destroy(name, state))
}

func TestAvailableIPAddressFull(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_available_ip_address"

	// neither a prefix nor a range
	typ := h.schemas.ResourceSchemas[name].ValueType()
	cfg := h.object(typ, nil)
	resp, err := h.server.ValidateResourceConfig(context.Background(), &tfprotov6.ValidateResourceConfigRequest{TypeName: name, Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	if !h.hasError(resp.Diagnostics) {
		t.Error("a configuration without prefix_id and ip_range_id was accepted")
	}
}

func TestResourceConfigErrors(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()

	// a name that is too long, and an attribute with a value that is not allowed
	for _, tt := range []struct {
		typeName string
		set      map[string]tftypes.Value
	}{
		{"netbox_tenant", map[string]tftypes.Value{"name": str(string(make([]byte, 101)))}},
		{"netbox_device", map[string]tftypes.Value{
			"name": str("d"), "device_type_id": num(1), "role_id": num(1), "site_id": num(1), "status": str("nonsense"),
		}},
		{"netbox_device", map[string]tftypes.Value{
			"name": str("d"), "device_type_id": num(1), "role_id": num(1), "site_id": num(1), "rack_face": str("front"),
		}},
	} {
		cfg := h.object(h.schemas.ResourceSchemas[tt.typeName].ValueType(), tt.set)
		resp, err := h.server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{TypeName: tt.typeName, Config: h.dynamic(cfg)})
		if err != nil {
			t.Fatal(err)
		}
		if !h.hasError(resp.Diagnostics) {
			t.Errorf("%s: an invalid configuration was accepted: %v", tt.typeName, tt.set)
		}
	}
}

func TestApiErrorsAreReported(t *testing.T) {
	t.Setenv("NETBOX_MAX_RETRIES", "0")
	fake, srv := newFakeNetBox()
	h := newHarness(t, srv.URL)
	_ = fake
	typ := h.schemas.ResourceSchemas["netbox_tenant"].ValueType()
	cfg := h.object(typ, map[string]tftypes.Value{"name": str("x")})
	null := tftypes.NewValue(typ, nil)
	plan, err := h.server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: "netbox_tenant", PriorState: h.dynamic(null), ProposedNewState: h.dynamic(cfg), Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}

	// the server goes away before the change is applied
	srv.Close()
	applied, err := h.server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "netbox_tenant", PriorState: h.dynamic(null), PlannedState: plan.PlannedState, Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	if !h.hasError(applied.Diagnostics) {
		t.Error("an unreachable NetBox did not fail the create")
	}

	state := h.object(typ, map[string]tftypes.Value{"id": str("1"), "name": str("x")})
	if rr := h.read("netbox_tenant", state); !h.hasError(rr.Diagnostics) {
		t.Error("an unreachable NetBox did not fail the read")
	}
	if !h.hasError(h.destroy("netbox_tenant", state)) {
		t.Error("an unreachable NetBox did not fail the delete")
	}
}

func mapOf(m map[string]string) tftypes.Value {
	vals := map[string]tftypes.Value{}
	for k, v := range m {
		vals[k] = str(v)
	}
	return tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, vals)
}

// A custom field that is no longer in the configuration is cleared in NetBox.
func TestRemovedCustomFieldIsCleared(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_tenant"

	state := h.apply(name, nullObject(), map[string]tftypes.Value{
		"name": str("cf"), "custom_fields": mapOf(map[string]string{"a": "1", "b": "2"}),
	})
	obj := fake.get("tenancy/tenants", int(numberAttr(t, state, "id")))

	state = h.apply(name, state, map[string]tftypes.Value{"name": str("cf"), "custom_fields": mapOf(map[string]string{"a": "1"})})
	cf, _ := obj["custom_fields"].(map[string]any)
	if v, present := cf["b"]; !present || v != nil {
		t.Errorf("b was not cleared: %v", cf)
	}
	if cf["a"] != "1" {
		t.Errorf("a = %v", cf["a"])
	}

	// not mentioning custom fields at all leaves them alone
	obj["custom_fields"] = map[string]any{"a": "1", "c": "3"}
	h.apply(name, state, map[string]tftypes.Value{"name": str("cf")})
	if cf, _ := obj["custom_fields"].(map[string]any); cf["a"] != "1" || cf["c"] != "3" {
		t.Errorf("custom fields were touched without being configured: %v", cf)
	}
}

// A tag with the name of a default tag is shown on an object that does not use default tags.
func TestNoDefaultTagsShowsAllTags(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	fake.put("extras/tags", map[string]any{"id": float64(1), "name": "managed", "slug": "managed"})

	h := newHarness(t, srv.URL)
	cfg := h.object(h.schemas.Provider.ValueType(), map[string]tftypes.Value{
		"server_url": str(srv.URL), "api_token": str("t"), "default_tags": setOf(tftypes.String, str("managed")),
	})
	resp, err := h.server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("configure", resp.Diagnostics)

	state := h.apply("netbox_tag", nullObject(), map[string]tftypes.Value{"name": str("plain")})
	obj := fake.get("extras/tags", int(numberAttr(t, state, "id")))
	obj["tags"] = []any{map[string]any{"name": "managed", "slug": "managed"}}

	got := h.decode(h.schemas.ResourceSchemas["netbox_tag"].ValueType(), h.read("netbox_tag", state).NewState)
	var tags []tftypes.Value
	if err := got["tags"].As(&tags); err != nil || len(tags) != 1 {
		t.Errorf("tags = %v, %v", got["tags"], err)
	}
}
