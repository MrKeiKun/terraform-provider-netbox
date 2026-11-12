package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func (f *fakeNetBox) put(collection string, obj map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.objects[collection] == nil {
		f.objects[collection] = map[int]map[string]any{}
	}
	id := int(obj["id"].(float64))
	f.objects[collection][id] = obj
	if id >= f.next {
		f.next = id + 1
	}
}

func (f *fakeNetBox) get(collection string, id int) map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.objects[collection][id]
}

func num(n int64) tftypes.Value { return tftypes.NewValue(tftypes.Number, n) }

func nullObject() tftypes.Value { return tftypes.NewValue(tftypes.Object{}, nil) }

func isGone(t *testing.T, h *harness, name string, rr *tfprotov6.ReadResourceResponse) bool {
	t.Helper()
	h.noErrors("read", rr.Diagnostics)
	if rr.NewState == nil {
		return true
	}
	v, err := rr.NewState.Unmarshal(h.schemas.ResourceSchemas[name].ValueType())
	if err != nil {
		t.Fatal(err)
	}
	return v.IsNull()
}

func TestAssignLifecycle(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()
	const name = "netbox_device_primary_ip"

	fake.put("dcim/devices", map[string]any{"id": float64(5), "name": "d"})
	fake.put("ipam/ip-addresses", map[string]any{"id": float64(10)})
	fake.put("ipam/ip-addresses", map[string]any{"id": float64(11)})

	state := h.apply(name, nullObject(), map[string]tftypes.Value{"device_id": num(5), "ip_address_id": num(10), "ip_address_version": num(4)})
	if got := stringAttr(t, state, "id"); got != "5" {
		t.Fatalf("id = %q", got)
	}
	if fake.get("dcim/devices", 5)["primary_ip4"] == nil {
		t.Error("the primary IP was not set on the device")
	}

	state = h.apply(name, state, map[string]tftypes.Value{"device_id": num(5), "ip_address_id": num(11), "ip_address_version": num(4)})

	// the assignment is read back from the device
	fake.get("dcim/devices", 5)["primary_ip4"] = map[string]any{"id": float64(11)}
	if isGone(t, h, name, h.read(name, state)) {
		t.Error("the assignment was dropped although the device still has it")
	}

	imp, err := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: "5"})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("import", imp.Diagnostics)
	if bad, _ := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: "x"}); bad == nil || !h.hasError(bad.Diagnostics) {
		t.Error("a bad import ID was accepted")
	}

	h.noErrors("destroy", h.destroy(name, state))

	// an assignment that is gone from the device is dropped from the state
	fake.get("dcim/devices", 5)["primary_ip4"] = nil
	if !isGone(t, h, name, h.read(name, state)) {
		t.Error("the state still holds the assignment")
	}

	// a device that does not exist
	typ := h.schemas.ResourceSchemas[name].ValueType()
	missing := h.object(typ, map[string]tftypes.Value{"id": str("99"), "device_id": num(99), "ip_address_id": num(10), "ip_address_version": num(4)})
	if !isGone(t, h, name, h.read(name, missing)) {
		t.Error("the state still holds an assignment on a missing device")
	}
}

func TestDeviceTagLifecycle(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()
	const name = "netbox_device_tag"

	fake.put("dcim/devices", map[string]any{"id": float64(7), "name": "d", "tags": []any{}})
	fake.put("extras/tags", map[string]any{"id": float64(1), "name": "blue", "slug": "blue"})
	fake.put("extras/tags", map[string]any{"id": float64(2), "name": "red", "slug": "red"})

	state := h.apply(name, nullObject(), map[string]tftypes.Value{"device_id": num(7), "tag": str("blue")})
	if got := stringAttr(t, state, "id"); got != "7:blue" {
		t.Fatalf("id = %q", got)
	}
	if tags, _ := fake.get("dcim/devices", 7)["tags"].([]any); len(tags) != 1 {
		t.Fatalf("tags on the device = %v", fake.get("dcim/devices", 7)["tags"])
	}

	// a second tag on the same device keeps the first
	state2 := h.apply(name, nullObject(), map[string]tftypes.Value{"device_id": num(7), "tag": str("red")})
	if tags, _ := fake.get("dcim/devices", 7)["tags"].([]any); len(tags) != 2 {
		t.Fatalf("tags on the device = %v", fake.get("dcim/devices", 7)["tags"])
	}

	if isGone(t, h, name, h.read(name, state)) {
		t.Error("the tag is on the device but was dropped")
	}

	imp, err := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: "7:blue"})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("import", imp.Diagnostics)
	for _, bad := range []string{"7", "x:blue"} {
		resp, _ := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: bad})
		if resp == nil || !h.hasError(resp.Diagnostics) {
			t.Errorf("import ID %q was accepted", bad)
		}
	}

	h.noErrors("destroy", h.destroy(name, state))
	if tags, _ := fake.get("dcim/devices", 7)["tags"].([]any); len(tags) != 1 {
		t.Errorf("tags after destroy = %v", fake.get("dcim/devices", 7)["tags"])
	}

	// the tag is gone from the device
	if !isGone(t, h, name, h.read(name, state)) {
		t.Error("the state still holds the removed tag")
	}
	h.noErrors("destroy second", h.destroy(name, state2))

	// an unknown tag fails
	typ := h.schemas.ResourceSchemas[name].ValueType()
	cfg := h.object(typ, map[string]tftypes.Value{"device_id": num(7), "tag": str("missing")})
	null := tftypes.NewValue(typ, nil)
	plan, err := h.server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: h.dynamic(null), ProposedNewState: h.dynamic(cfg), Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	applied, err := h.server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: h.dynamic(null), PlannedState: plan.PlannedState, Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	if !h.hasError(applied.Diagnostics) {
		t.Error("an unknown tag was accepted")
	}
}

func TestPluralDataSource(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()

	for i, n := range []string{"alpha", "beta", "gamma"} {
		fake.put("tenancy/tenants", map[string]any{"id": float64(i + 1), "name": n, "slug": n})
	}

	typ := h.schemas.DataSourceSchemas["netbox_tenants"].ValueType()
	read := func(set map[string]tftypes.Value) []tftypes.Value {
		resp, err := h.server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{TypeName: "netbox_tenants", Config: h.dynamic(h.object(typ, set))})
		if err != nil {
			t.Fatal(err)
		}
		h.noErrors("read", resp.Diagnostics)
		var list []tftypes.Value
		if err := h.decode(typ, resp.State)["tenants"].As(&list); err != nil {
			t.Fatal(err)
		}
		return list
	}

	if got := len(read(nil)); got != 3 {
		t.Errorf("all tenants = %d, want 3", got)
	}

	// a filter narrows the result
	filterType := typ.(tftypes.Object).AttributeTypes["filter"]
	var elem tftypes.Type
	switch ft := filterType.(type) {
	case tftypes.Set:
		elem = ft.ElementType
	case tftypes.List:
		elem = ft.ElementType
	default:
		t.Fatalf("unexpected filter type %T", filterType)
	}
	item := tftypes.NewValue(elem, map[string]tftypes.Value{"name": str("name"), "value": str("beta")})
	filter := tftypes.NewValue(filterType, []tftypes.Value{item})
	if got := len(read(map[string]tftypes.Value{"filter": filter})); got != 1 {
		t.Errorf("filtered tenants = %d, want 1", got)
	}

	if got := len(read(map[string]tftypes.Value{"limit": num(2)})); got > 2 {
		t.Errorf("limited tenants = %d, want at most 2", got)
	}
}

// Deleting an assignment leaves the field alone when it points at another object by now.
func TestAssignDeleteKeepsChangedField(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_device_primary_ip"

	fake.put("dcim/devices", map[string]any{"id": float64(5), "name": "d"})
	state := h.apply(name, nullObject(), map[string]tftypes.Value{"device_id": num(5), "ip_address_id": num(10), "ip_address_version": num(4)})

	fake.get("dcim/devices", 5)["primary_ip4"] = map[string]any{"id": float64(99)}
	h.noErrors("destroy", h.destroy(name, state))
	if ref, _ := fake.get("dcim/devices", 5)["primary_ip4"].(map[string]any); ref == nil || ref["id"] != float64(99) {
		t.Errorf("an assignment made elsewhere was cleared: %v", fake.get("dcim/devices", 5)["primary_ip4"])
	}

	// when it still points at our IP, it is cleared
	state = h.apply(name, nullObject(), map[string]tftypes.Value{"device_id": num(5), "ip_address_id": num(10), "ip_address_version": num(4)})
	fake.get("dcim/devices", 5)["primary_ip4"] = map[string]any{"id": float64(10)}
	h.noErrors("destroy", h.destroy(name, state))
	if fake.get("dcim/devices", 5)["primary_ip4"] != nil {
		t.Errorf("the assignment was not cleared: %v", fake.get("dcim/devices", 5)["primary_ip4"])
	}
}
