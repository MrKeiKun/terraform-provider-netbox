package fwprovider

import (
	"context"
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPortMappingHooksWithSingle(t *testing.T) {
	s := &fwresource.Spec{}
	portMappingHooks(s, "rear_ports", "rear_port", true)
	ctx := context.Background()

	plan := fwresource.Values{
		"rear_port_id":       types.Int64Value(4),
		"rear_port_position": types.Int64Value(2),
		"rear_ports":         types.ListNull(types.StringType),
	}
	body := map[string]any{"rear_port_position": int64(2), "rear_port": int64(4)}
	if err := s.ModifyBody(ctx, plan, nil, body); err != nil {
		t.Fatal(err)
	}
	maps, _ := body["rear_ports"].([]any)
	if len(maps) != 1 {
		t.Fatalf("mappings = %v", body["rear_ports"])
	}
	if m := maps[0].(map[string]any); m["rear_port"] != int64(4) || m["rear_port_position"] != int64(2) {
		t.Errorf("mapping = %v", m)
	}
	if _, ok := body["rear_port_position"]; ok {
		t.Error("the single position was sent along with the mapping")
	}

	// the position defaults to 1
	plan["rear_port_position"] = types.Int64Value(0)
	body = map[string]any{}
	_ = s.ModifyBody(ctx, plan, nil, body)
	if m := body["rear_ports"].([]any)[0].(map[string]any); m["rear_port_position"] != int64(1) {
		t.Errorf("default position = %v", m["rear_port_position"])
	}

	// an update that changes nothing sends no mapping
	body = map[string]any{"rear_port": int64(4)}
	_ = s.ModifyBody(ctx, plan, plan, body)
	if _, ok := body["rear_ports"]; ok {
		t.Error("an unchanged mapping was sent")
	}

	// no port, nothing to send
	body = map[string]any{}
	_ = s.ModifyBody(ctx, fwresource.Values{"rear_port_id": types.Int64Null()}, nil, body)
	if len(body) != 0 {
		t.Errorf("body = %v", body)
	}

	// the single port is filled from the mapping NetBox returns
	out := fwresource.Values{}
	obj := map[string]any{"rear_ports": []any{map[string]any{"rear_port": map[string]any{"id": float64(9)}, "rear_port_position": float64(3)}}}
	if err := s.ModifyState(ctx, obj, nil, out); err != nil {
		t.Fatal(err)
	}
	if out["rear_port_id"] != types.Int64Value(9) || out["rear_port_position"] != types.Int64Value(3) {
		t.Errorf("state = %v", out)
	}
	out = fwresource.Values{}
	_ = s.ModifyState(ctx, map[string]any{"rear_port": float64(1)}, nil, out)
	_ = s.ModifyState(ctx, map[string]any{}, nil, out)
	if len(out) != 0 {
		t.Errorf("state = %v", out)
	}
}

func TestPortMappingHooksWithoutSingle(t *testing.T) {
	s := &fwresource.Spec{}
	portMappingHooks(s, "front_ports", "front_port", false)
	list := types.ListNull(types.StringType)
	plan := fwresource.Values{"front_ports": list}

	body := map[string]any{"front_ports": []any{}}
	_ = s.ModifyBody(context.Background(), plan, plan, body)
	if _, ok := body["front_ports"]; ok {
		t.Error("an unchanged mapping was sent")
	}
	body = map[string]any{"front_ports": []any{}}
	_ = s.ModifyBody(context.Background(), plan, nil, body)
	if _, ok := body["front_ports"]; !ok {
		t.Error("the mapping was dropped on create")
	}
	if s.ModifyState != nil {
		t.Error("there is no single port to fill")
	}
}

func TestSmallHelpers(t *testing.T) {
	if portInt(float64(3)) != 3 || portInt(map[string]any{"id": float64(5)}) != 5 || portInt("x") != 0 {
		t.Error("portInt")
	}
	one, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("a")})
	set, _ := types.SetValue(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	if listLen(one) != 1 || listLen(set) != 2 || listLen(types.StringValue("x")) != 0 {
		t.Error("listLen")
	}
	cfg := fwresource.Values{"n": types.Int64Value(4), "s": types.StringValue("v"), "null": types.Int64Null()}
	if intCfg(cfg, "n") != 4 || intCfg(cfg, "null") != 0 || intCfg(cfg, "missing") != 0 {
		t.Error("intCfg")
	}
	if strCfg(cfg, "s") != "v" || strCfg(cfg, "n") != "" || strCfg(cfg, "missing") != "" {
		t.Error("strCfg")
	}
	if got := netboxTime("2030-01-01T01:00:00+01:00"); got != "2030-01-01T00:00:00.000Z" {
		t.Errorf("netboxTime = %v", got)
	}
	if got := netboxTime("not a time"); got != "not a time" {
		t.Errorf("netboxTime of garbage = %v", got)
	}
	if boolValue(fwresource.Values{"b": types.BoolValue(true)}, "b") != true || boolValue(fwresource.Values{}, "b") {
		t.Error("boolValue")
	}
	if intValue(fwresource.Values{"n": types.Int64Value(2)}, "n") != 2 || stringValue(fwresource.Values{"s": types.StringValue("x")}, "s") != "x" {
		t.Error("intValue and stringValue")
	}
}

func TestAvailablePrefixLifecycle(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()
	const name = "netbox_available_prefix"

	fake.put("ipam/prefixes", map[string]any{"id": float64(1), "prefix": "10.1.0.0/16"})

	state := h.apply(name, nullObject(), map[string]tftypes.Value{
		"parent_prefix_id": num(1), "prefix_length": num(24), "status": str("active"),
	})
	if got := stringAttr(t, state, "prefix"); got == "" {
		t.Error("no prefix was allocated")
	}
	prefix := stringAttr(t, state, "prefix")
	state = h.apply(name, state, map[string]tftypes.Value{
		"parent_prefix_id": num(1), "prefix_length": num(24), "status": str("reserved"),
	})
	if got := stringAttr(t, state, "prefix"); got != prefix {
		t.Errorf("prefix changed from %s to %s", prefix, got)
	}

	id := stringAttr(t, state, "id")
	imp, err := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: "1 " + id + " 24"})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("import", imp.Diagnostics)
	for _, bad := range []string{"1", "x 2 24", "1 y 24", "1 2 z", "1  24"} {
		resp, _ := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: bad})
		if resp == nil || !h.hasError(resp.Diagnostics) {
			t.Errorf("import ID %q was accepted", bad)
		}
	}
	h.noErrors("destroy", h.destroy(name, state))
}

func TestAvailableVLANLifecycle(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_available_vlan"

	fake.put("ipam/vlan-groups", map[string]any{"id": float64(1), "name": "g"})
	state := h.apply(name, nullObject(), map[string]tftypes.Value{"name": str("v"), "group_id": num(1)})
	if numberAttr(t, state, "id") == 0 {
		t.Fatal("no VLAN was allocated")
	}
	state = h.apply(name, state, map[string]tftypes.Value{"name": str("renamed"), "group_id": num(1)})
	if got := stringAttr(t, state, "name"); got != "renamed" {
		t.Errorf("name = %q", got)
	}

	// without a group there is nothing to allocate from
	typ := h.schemas.ResourceSchemas[name].ValueType()
	cfg := h.object(typ, map[string]tftypes.Value{"name": str("lonely")})
	null := tftypes.NewValue(typ, nil)
	plan, err := h.server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: h.dynamic(null), ProposedNewState: h.dynamic(cfg), Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	applied, err := h.server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: h.dynamic(null), PlannedState: plan.PlannedState, Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	if !h.hasError(applied.Diagnostics) {
		t.Error("a VLAN without a group was allocated")
	}
	h.noErrors("destroy", h.destroy(name, state))
}

func TestAvailableIPAddressImport(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()
	const name = "netbox_available_ip_address"

	fake.put("ipam/prefixes", map[string]any{"id": float64(1), "prefix": "10.0.0.0/16"})
	fake.put("ipam/prefixes", map[string]any{"id": float64(2), "prefix": "10.0.0.0/24"})
	fake.put("ipam/ip-addresses", map[string]any{"id": float64(50), "address": "10.0.0.5/24", "status": map[string]any{"value": "active"}})

	imp, err := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: "50"})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("import", imp.Diagnostics)
	rr, err := h.server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: imp.ImportedResources[0].State})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("read", rr.Diagnostics)
	got := h.decode(h.schemas.ResourceSchemas[name].ValueType(), rr.NewState)
	if got["prefix_id"].IsNull() {
		t.Error("the parent prefix was not found after the import")
	}
}

func TestDeviceWithVirtualChassis(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_device"

	fake.put("dcim/virtual-chassis", map[string]any{"id": float64(1), "name": "vc"})
	set := func(extra map[string]tftypes.Value) map[string]tftypes.Value {
		m := map[string]tftypes.Value{"name": str("sw1"), "device_type_id": num(1), "role_id": num(1), "site_id": num(1), "virtual_chassis_id": num(1)}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}

	state := h.apply(name, nullObject(), set(map[string]tftypes.Value{"virtual_chassis_master": tftypes.NewValue(tftypes.Bool, true)}))
	id := numberAttr(t, state, "id")
	vc := fake.get("dcim/virtual-chassis", 1)
	if vc["master"] != float64(id) && vc["master"] != id {
		t.Errorf("master of the chassis = %v, want %d", vc["master"], id)
	}

	// NetBox reports the chassis and its master on the device
	dev := fake.get("dcim/devices", int(id))
	dev["virtual_chassis"] = map[string]any{"id": float64(1), "master": map[string]any{"id": float64(id)}}
	state = h.readState(name, state)

	// no longer the master
	state = h.apply(name, state, set(map[string]tftypes.Value{"virtual_chassis_master": tftypes.NewValue(tftypes.Bool, false)}))
	if vc["master"] != nil {
		t.Errorf("master after the change = %v", vc["master"])
	}

	// the master is released before the device is deleted
	state = h.apply(name, state, set(map[string]tftypes.Value{"virtual_chassis_master": tftypes.NewValue(tftypes.Bool, true)}))
	dev["virtual_chassis"] = map[string]any{"id": float64(1), "master": map[string]any{"id": float64(id)}}
	state = h.readState(name, state)
	h.noErrors("destroy", h.destroy(name, state))
	if vc["master"] != nil {
		t.Errorf("master after the delete = %v", vc["master"])
	}

	// a device with a rack keeps the location NetBox derives from it
	state = h.apply(name, nullObject(), map[string]tftypes.Value{
		"name": str("sw2"), "device_type_id": num(1), "role_id": num(1), "site_id": num(1), "rack_id": num(3),
	})
	h.noErrors("destroy second", h.destroy(name, state))
}

func TestVirtualMachineLifecycle(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_virtual_machine"

	state := h.apply(name, nullObject(), map[string]tftypes.Value{
		"name": str("vm"), "cluster_id": num(1), "memory_mb": num(2048), "vcpus": tftypes.NewValue(tftypes.Number, 2.5),
	})
	obj := fake.get("virtualization/virtual-machines", int(numberAttr(t, state, "id")))
	if obj["memory"] != float64(2048) {
		t.Errorf("memory sent = %v", obj["memory"])
	}
	// a put sends the primary IPs as they are
	obj["primary_ip4"] = map[string]any{"id": float64(7)}
	state = h.apply(name, h.readState(name, state), map[string]tftypes.Value{
		"name": str("vm2"), "cluster_id": num(1), "memory_mb": num(1024),
	})
	if obj["primary_ip4"] != float64(7) {
		t.Errorf("primary_ip4 after the update = %v", obj["primary_ip4"])
	}
	h.noErrors("destroy", h.destroy(name, state))
}

func (h *harness) readState(typeName string, state tftypes.Value) tftypes.Value {
	h.t.Helper()
	rr := h.read(typeName, state)
	h.noErrors("read", rr.Diagnostics)
	v, err := rr.NewState.Unmarshal(h.schemas.ResourceSchemas[typeName].ValueType())
	if err != nil {
		h.t.Fatal(err)
	}
	return v
}

func TestCustomDataSources(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()

	fake.put("ipam/ip-addresses", map[string]any{
		"id": float64(3), "address": "10.0.0.3/24", "created": "2030-01-01T01:00:00+01:00", "last_updated": "2030-01-01T01:00:00+01:00",
		"family": map[string]any{"value": float64(4), "label": "IPv4"}, "status": map[string]any{"value": "active"},
	})
	fake.put("dcim/rack-types", map[string]any{"id": float64(4), "model": "m1", "slug": "m1"})
	fake.put("dcim/rack-types", map[string]any{"id": float64(5), "model": "m2", "slug": "m2"})
	fake.put("extras/notification-groups", map[string]any{"id": float64(6), "name": "ops"})
	fake.put("extras/notification-groups", map[string]any{"id": float64(7), "name": "ops-extra"})

	read := func(name string, set map[string]tftypes.Value) *tfprotov6.ReadDataSourceResponse {
		typ := h.schemas.DataSourceSchemas[name].ValueType()
		resp, err := h.server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{TypeName: name, Config: h.dynamic(h.object(typ, set))})
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}

	if resp := read("netbox_ip_address", map[string]tftypes.Value{"id": num(3)}); h.hasError(resp.Diagnostics) {
		t.Errorf("ip address: %v", resp.Diagnostics[0].Detail)
	}
	if resp := read("netbox_ip_address", map[string]tftypes.Value{"id": num(99)}); !h.hasError(resp.Diagnostics) {
		t.Error("a missing IP address did not fail")
	}

	if resp := read("netbox_rack_type", map[string]tftypes.Value{"model": str("m2")}); h.hasError(resp.Diagnostics) {
		t.Errorf("rack type: %v", resp.Diagnostics[0].Detail)
	}
	if resp := read("netbox_rack_type", map[string]tftypes.Value{"slug": str("m1")}); h.hasError(resp.Diagnostics) {
		t.Errorf("rack type by slug: %v", resp.Diagnostics[0].Detail)
	}
	if resp := read("netbox_rack_type", nil); !h.hasError(resp.Diagnostics) {
		t.Error("a rack type search without model or slug did not fail")
	}
	if resp := read("netbox_rack_type", map[string]tftypes.Value{"model": str("none")}); !h.hasError(resp.Diagnostics) {
		t.Error("a missing rack type did not fail")
	}

	if resp := read("netbox_notification_group", map[string]tftypes.Value{"name": str("ops")}); h.hasError(resp.Diagnostics) {
		t.Errorf("notification group: %v", resp.Diagnostics[0].Detail)
	}
	if resp := read("netbox_notification_group", map[string]tftypes.Value{"name": str("nothing")}); !h.hasError(resp.Diagnostics) {
		t.Error("a missing notification group did not fail")
	}
}

// After an import the parent is not in the configuration's attribute, and that must not replace the address.
func TestImportedAvailableIPIsNotReplaced(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()
	const name = "netbox_available_ip_address"

	fake.put("ipam/prefixes", map[string]any{"id": float64(1), "prefix": "10.0.0.0/24"})
	fake.put("ipam/ip-addresses", map[string]any{"id": float64(50), "address": "10.0.0.5/24"})

	imp, err := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: "50"})
	if err != nil {
		t.Fatal(err)
	}
	rr, _ := h.server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: imp.ImportedResources[0].State})
	typ := h.schemas.ResourceSchemas[name].ValueType()
	prior, _ := rr.NewState.Unmarshal(typ)

	// the configuration names an IP range, the state holds the prefix found by the import
	cfg := h.object(typ, map[string]tftypes.Value{"ip_range_id": num(7)})
	plan, err := h.server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName: name, PriorState: h.dynamic(prior), ProposedNewState: h.dynamic(h.proposed(name, typ, prior, cfg)), Config: h.dynamic(cfg),
	})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("plan", plan.Diagnostics)
	if len(plan.RequiresReplace) != 0 {
		t.Errorf("the address would be replaced: %v", plan.RequiresReplace)
	}

	// a change from one prefix to another still replaces it
	var m map[string]tftypes.Value
	_ = prior.As(&m)
	m["prefix_id"] = num(1)
	prior = tftypes.NewValue(typ, m)
	cfg = h.object(typ, map[string]tftypes.Value{"prefix_id": num(2)})
	plan, err = h.server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName: name, PriorState: h.dynamic(prior), ProposedNewState: h.dynamic(h.proposed(name, typ, prior, cfg)), Config: h.dynamic(cfg),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.RequiresReplace) == 0 {
		t.Error("a different prefix did not replace the address")
	}
}
