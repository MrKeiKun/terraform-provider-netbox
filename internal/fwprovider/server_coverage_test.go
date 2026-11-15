package fwprovider

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// The fake NetBox reports version 4.4.0, so objects and attributes of newer versions are refused.
func TestVersionGates(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()
	_ = fake

	// the plan is refused, so the diagnostics come from the plan and not from an apply
	plan := func(typeName string, set map[string]tftypes.Value) []*tfprotov6.Diagnostic {
		typ := h.schemas.ResourceSchemas[typeName].ValueType()
		cfg := h.object(typ, set)
		prior := tftypes.NewValue(typ, nil)
		resp, err := h.server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
			TypeName: typeName, PriorState: h.dynamic(prior), ProposedNewState: h.dynamic(h.proposed(typeName, typ, prior, cfg)), Config: h.dynamic(cfg),
		})
		if err != nil {
			t.Fatal(err)
		}
		return resp.Diagnostics
	}
	h.wantError("an object type that needs a newer NetBox", plan("netbox_cable_bundle", map[string]tftypes.Value{"name": str("b")}))
	h.wantError("an attribute that needs a newer NetBox", plan("netbox_site", map[string]tftypes.Value{"name": str("s"), "slug": str("s"), "owner_id": num(1)}))

	typ := h.schemas.DataSourceSchemas["netbox_cable_bundle"].ValueType()
	resp, err := h.server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{TypeName: "netbox_cable_bundle", Config: h.dynamic(h.object(typ, map[string]tftypes.Value{"name": str("b")}))})
	if err != nil {
		t.Fatal(err)
	}
	h.wantError("a data source that needs a newer NetBox", resp.Diagnostics)
}

func TestDataSourceFilters(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()

	fake.put("tenancy/tenants", map[string]any{"id": float64(1), "name": "a", "slug": "a", "cf_team": "red"})
	fake.put("tenancy/tenants", map[string]any{"id": float64(2), "name": "b", "slug": "b"})
	fake.put("tenancy/tenants", map[string]any{"id": float64(3), "name": "twin", "slug": "twin-1"})
	fake.put("tenancy/tenants", map[string]any{"id": float64(4), "name": "twin", "slug": "twin-2"})
	fake.put("dcim/regions", map[string]any{"id": float64(7), "name": "north", "slug": "north"})
	fake.put("virtualization/clusters", map[string]any{"id": float64(1), "name": "prod-1"})
	fake.put("virtualization/clusters", map[string]any{"id": float64(2), "name": "test-1"})

	read := func(name string, set map[string]tftypes.Value) *tfprotov6.ReadDataSourceResponse {
		typ := h.schemas.DataSourceSchemas[name].ValueType()
		resp, err := h.server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{TypeName: name, Config: h.dynamic(h.object(typ, set))})
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}
	// filters builds the filter blocks. Each one is a list of attribute name and value pairs, the other
	// attributes of the block stay null.
	filters := func(name string, blocks ...map[string]string) tftypes.Value {
		ft := h.schemas.DataSourceSchemas[name].ValueType().(tftypes.Object).AttributeTypes["filter"]
		var elem tftypes.Type
		switch x := ft.(type) {
		case tftypes.Set:
			elem = x.ElementType
		case tftypes.List:
			elem = x.ElementType
		}
		var vals []tftypes.Value
		for _, block := range blocks {
			attrs := map[string]tftypes.Value{}
			for attr, typ := range elem.(tftypes.Object).AttributeTypes {
				if v, ok := block[attr]; ok {
					attrs[attr] = str(v)
				} else {
					attrs[attr] = tftypes.NewValue(typ, nil)
				}
			}
			vals = append(vals, tftypes.NewValue(elem, attrs))
		}
		return tftypes.NewValue(ft, vals)
	}
	count := func(resp *tfprotov6.ReadDataSourceResponse, name, field string) int {
		out := h.decode(h.schemas.DataSourceSchemas[name].ValueType(), resp.State)
		var items []tftypes.Value
		if err := out[field].As(&items); err != nil {
			t.Fatalf("%s: %v", field, err)
		}
		return len(items)
	}

	// filters, a limit and an empty value
	resp := read("netbox_tenants", map[string]tftypes.Value{"filter": filters("netbox_tenants", map[string]string{"name": "name", "value": "a"})})
	h.noErrors("a filtered list", resp.Diagnostics)
	if got := count(resp, "netbox_tenants", "tenants"); got != 1 {
		t.Errorf("filtered list has %d tenants, want 1", got)
	}
	resp = read("netbox_tenants", map[string]tftypes.Value{"filter": filters("netbox_tenants", map[string]string{"name": "name", "value": ""})})
	h.noErrors("an empty filter value", resp.Diagnostics)
	if got := count(resp, "netbox_tenants", "tenants"); got != 4 {
		t.Errorf("list with an empty filter has %d tenants, want 4", got)
	}
	resp = read("netbox_tenants", map[string]tftypes.Value{"limit": num(2)})
	h.noErrors("a list with a limit", resp.Diagnostics)
	if got := count(resp, "netbox_tenants", "tenants"); got != 2 {
		t.Errorf("list with a limit has %d tenants, want 2", got)
	}

	// the name regular expression keeps the matching results
	resp = read("netbox_clusters", map[string]tftypes.Value{"name_regex": str("^prod-")})
	h.noErrors("a name regular expression", resp.Diagnostics)
	if got := count(resp, "netbox_clusters", "clusters"); got != 1 {
		t.Errorf("the regular expression kept %d clusters, want 1", got)
	}

	// custom fields filter a single lookup
	h.noErrors("a lookup by custom field", read("netbox_tenant", map[string]tftypes.Value{"custom_fields": mapOf(map[string]string{"team": "red"})}).Diagnostics)

	// the filter block of a single lookup
	h.noErrors("a lookup with a filter block", read("netbox_region", map[string]tftypes.Value{"filter": filters("netbox_region", map[string]string{"name": "north"})}).Diagnostics)

	// more than one match, and none
	h.wantError("a lookup that matches more than one object", read("netbox_tenant", map[string]tftypes.Value{"name": str("twin")}).Diagnostics)
	h.wantError("a lookup that matches nothing", read("netbox_tenant", map[string]tftypes.Value{"name": str("nobody")}).Diagnostics)
}

func TestAvailableASNLifecycle(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()
	const name = "netbox_available_asn"

	state := h.apply(name, nullObject(), map[string]tftypes.Value{"asn_range_id": num(5)})
	if numberAttr(t, state, "id") == 0 || bigAttr(t, state, "asn") == 0 {
		t.Fatalf("no AS number was allocated: %v", state)
	}
	state = h.apply(name, state, map[string]tftypes.Value{"asn_range_id": num(5), "description": str("renamed")})
	if got := stringAttr(t, state, "description"); got != "renamed" {
		t.Errorf("description = %q", got)
	}

	// a range of 0 has nothing to allocate from
	typ := h.schemas.ResourceSchemas[name].ValueType()
	cfg := h.object(typ, map[string]tftypes.Value{"asn_range_id": num(0)})
	null := tftypes.NewValue(typ, nil)
	plan, err := h.server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: h.dynamic(null), ProposedNewState: h.dynamic(cfg), Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	applied, err := h.server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: h.dynamic(null), PlannedState: plan.PlannedState, Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	h.wantError("an AS number without a range", applied.Diagnostics)

	// import needs the range and the AS number
	imp, err := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: "5 12"})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("import", imp.Diagnostics)
	for _, bad := range []string{"5", "x 12", "5 y", " 12", "5 "} {
		resp, err := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: bad})
		if err != nil {
			t.Fatal(err)
		}
		if !h.hasError(resp.Diagnostics) {
			t.Errorf("the import ID %q was accepted", bad)
		}
	}
	h.noErrors("destroy", h.destroy(name, state))
}

func TestAssignErrorPaths(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_device_interface_primary_mac_address"

	fake.put("dcim/interfaces", map[string]any{"id": float64(4), "name": "eth0", "tagged_vlans": []any{map[string]any{"id": float64(9)}}, "untagged_vlan": map[string]any{"id": float64(8)}})
	fake.put("dcim/mac-addresses", map[string]any{"id": float64(2), "mac_address": "aa:bb:cc:dd:ee:ff"})
	set := map[string]tftypes.Value{"interface_id": num(4), "mac_address_id": num(2)}

	// the VLANs of the interface are kept, so the interface is read first
	fake.failOn("GET", "dcim/interfaces", 500)
	diags, _ := h.tryApply(name, nullObject(), set)
	h.wantError("an assignment whose interface cannot be read", diags)
	fake.clearFailures()

	state := h.apply(name, nullObject(), set)
	fake.get("dcim/interfaces", 4)["primary_mac_address"] = map[string]any{"id": float64(2)}

	// an update that NetBox rejects
	fake.put("dcim/mac-addresses", map[string]any{"id": float64(3), "mac_address": "11:22:33:44:55:66"})
	fake.failOn("PATCH", "dcim/interfaces", 500)
	diags, _ = h.tryApply(name, state, map[string]tftypes.Value{"interface_id": num(4), "mac_address_id": num(3)})
	h.wantError("an update that NetBox rejects", diags)
	fake.clearFailures()

	// a read that NetBox rejects
	fake.failOn("GET", "dcim/interfaces", 500)
	h.wantError("a read that NetBox rejects", h.read(name, state).Diagnostics)
	h.wantError("a delete that cannot read the interface", h.destroy(name, state))
	fake.clearFailures()
}

// bigAttr returns an attribute that is a number in the schema.
func bigAttr(t *testing.T, state tftypes.Value, name string) int64 {
	t.Helper()
	var f big.Float
	if err := attrOf(t, state, name).As(&f); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	n, _ := f.Int64()
	return n
}

func TestResourceWriteErrors(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_contact_role"
	set := func(desc string) map[string]tftypes.Value {
		return map[string]tftypes.Value{"name": str("t"), "slug": str("t"), "description": str(desc)}
	}

	// the response to the write is the state, so a failing read does not matter
	fake.failOn("GET", "tenancy/contact-roles", 500)
	diags, _ := h.tryApply(name, nullObject(), set("a"))
	h.noErrors("a create does not read the object back", diags)
	fake.clearFailures()

	state := h.apply(name, nullObject(), set("a"))

	fake.failOn("PATCH", "tenancy/contact-roles", 500)
	diags, _ = h.tryApply(name, state, set("b"))
	h.wantError("an update that NetBox rejects", diags)
	fake.clearFailures()

	fake.failOn("GET", "tenancy/contact-roles", 500)
	diags, _ = h.tryApply(name, state, set("c"))
	h.noErrors("an update does not read the object back", diags)
	h.wantError("a read that NetBox rejects", h.read(name, state).Diagnostics)
	fake.clearFailures()
}

func TestDataSourceQueryFeatures(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()

	fake.put("virtualization/clusters", map[string]any{"id": float64(1), "name": "c1", "scope_type": "dcim.site", "scope_id": float64(5), "custom_fields": map[string]any{"team": "red"}, "cf_team": "red"})
	fake.put("circuits/provider-networks", map[string]any{"id": float64(2), "name": "pn", "custom_fields": map[string]any{"zone": "z1"}, "cf_zone": "z1"})
	fake.put("virtualization/virtual-disks", map[string]any{"id": float64(3), "name": "d1", "size": float64(10)})

	read := func(name string, set map[string]tftypes.Value) []*tfprotov6.Diagnostic {
		typ := h.schemas.DataSourceSchemas[name].ValueType()
		resp, err := h.server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{TypeName: name, Config: h.dynamic(h.object(typ, set))})
		if err != nil {
			t.Fatal(err)
		}
		return resp.Diagnostics
	}
	listOf := func(vals ...string) tftypes.Value {
		var out []tftypes.Value
		for _, v := range vals {
			out = append(out, str(v))
		}
		return tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, out)
	}

	// custom fields filter a lookup, and are returned with the object
	h.noErrors("a cluster by custom field", read("netbox_cluster", map[string]tftypes.Value{"custom_fields": mapOf(map[string]string{"team": "red"})}))
	h.noErrors("a provider network by custom field", read("netbox_circuit_provider_network", map[string]tftypes.Value{"custom_fields": mapOf(map[string]string{"zone": "z1"})}))
	h.noErrors("a cluster without custom field filter", read("netbox_cluster", map[string]tftypes.Value{"name": str("c1")}))

	// custom fields filter a list, and a value with commas becomes several values of the filter
	h.noErrors("clusters by custom field", read("netbox_clusters", map[string]tftypes.Value{"custom_fields": mapOf(map[string]string{"team": "red"})}))

	// the ordering of a list of virtual disks
	h.noErrors("virtual disks with an ordering", read("netbox_virtual_disks", map[string]tftypes.Value{"ordering": listOf("name", "-size")}))
	h.noErrors("virtual disks without an ordering", read("netbox_virtual_disks", nil))

	// a list of clusters carries the object a cluster is scoped to
	h.noErrors("clusters with their scope", read("netbox_clusters", nil))
}

func TestInventoryAccessors(t *testing.T) {
	if len(GeneratedFilters()) == 0 || len(ExtraDataSourceAttrs()) == 0 || len(ExtraResourceAttrs()) == 0 {
		t.Error("an accessor of the generated tables returned nothing")
	}
}

func TestCustomFieldChoiceSetNeedsPairs(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_custom_field_choice_set"
	typ := h.schemas.ResourceSchemas[name].ValueType().(tftypes.Object)
	listT := typ.AttributeTypes["extra_choices"]
	inner := listT.(tftypes.List).ElementType
	pair := func(vals ...string) tftypes.Value {
		var elems []tftypes.Value
		for _, v := range vals {
			elems = append(elems, str(v))
		}
		return tftypes.NewValue(inner, elems)
	}

	good := tftypes.NewValue(listT, []tftypes.Value{pair("a", "A")})
	state := h.apply(name, nullObject(), map[string]tftypes.Value{"name": str("c"), "extra_choices": good})
	if stringAttr(t, state, "name") != "c" {
		t.Error("a choice set with pairs was not created")
	}

	bad := tftypes.NewValue(listT, []tftypes.Value{pair("only one")})
	diags, _ := h.tryApply(name, nullObject(), map[string]tftypes.Value{"name": str("d"), "extra_choices": bad})
	h.wantError("a choice that is not a pair", diags)
}

// A tag is looked up once, whatever the number of objects that use it.
func TestTagLookupsAreKept(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_tenant"
	fake.put("extras/tags", map[string]any{"id": float64(1), "name": "blue", "slug": "blue"})
	tags := setOf(tftypes.String, str("blue"))

	lookups := func() int {
		n := 0
		for _, c := range fake.calls {
			if c == "GET /api/extras/tags/" {
				n++
			}
		}
		return n
	}
	for i, slug := range []string{"a", "b", "c"} {
		h.apply(name, nullObject(), map[string]tftypes.Value{"name": str(slug), "slug": str(slug), "tags": tags})
		if got := lookups(); got != 1 {
			t.Fatalf("after %d objects with the tag NetBox was asked %d times, want 1", i+1, got)
		}
	}

	// a tag that is created is looked up when it is first used
	red := setOf(tftypes.String, str("red"))
	state := h.apply("netbox_tag", nullObject(), map[string]tftypes.Value{"name": str("red"), "slug": str("red")})
	before := lookups()
	h.apply(name, nullObject(), map[string]tftypes.Value{"name": str("d"), "slug": str("d"), "tags": red})
	if got := lookups(); got != before+1 {
		t.Errorf("a tag that was created was not looked up: %d lookups, want %d", got, before+1)
	}

	// the tag is renamed, so the new name is looked up
	state = h.apply("netbox_tag", state, map[string]tftypes.Value{"name": str("green"), "slug": str("red")})
	green := setOf(tftypes.String, str("green"))
	before = lookups()
	h.apply(name, nullObject(), map[string]tftypes.Value{"name": str("e"), "slug": str("e"), "tags": green})
	if got := lookups(); got != before+1 {
		t.Errorf("a tag that was renamed was not looked up again: %d lookups, want %d", got, before+1)
	}

	h.noErrors("destroy of the tag", h.destroy("netbox_tag", state))
}
