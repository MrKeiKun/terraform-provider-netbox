package fwprovider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// tryApply plans and applies a change and returns what the apply reported, instead of stopping the
// test on an error.
func (h *harness) tryApply(typeName string, prior tftypes.Value, set map[string]tftypes.Value) ([]*tfprotov6.Diagnostic, tftypes.Value) {
	h.t.Helper()
	ctx := context.Background()
	typ := h.schemas.ResourceSchemas[typeName].ValueType()
	cfg := h.object(typ, set)
	if prior.IsNull() {
		prior = tftypes.NewValue(typ, nil)
	}
	plan, err := h.server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName: typeName, PriorState: h.dynamic(prior), ProposedNewState: h.dynamic(h.proposed(typeName, typ, prior, cfg)), Config: h.dynamic(cfg),
	})
	if err != nil {
		h.t.Fatal(err)
	}
	h.noErrors("plan", plan.Diagnostics)
	applied, err := h.server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{
		TypeName: typeName, PriorState: h.dynamic(prior), PlannedState: plan.PlannedState, Config: h.dynamic(cfg), PlannedPrivate: plan.PlannedPrivate,
	})
	if err != nil {
		h.t.Fatal(err)
	}
	state := tftypes.NewValue(typ, nil)
	if applied.NewState != nil {
		if v, err := applied.NewState.Unmarshal(typ); err == nil {
			state = v
		}
	}
	return applied.Diagnostics, state
}

func (h *harness) wantError(what string, diags []*tfprotov6.Diagnostic) {
	h.t.Helper()
	if !h.hasError(diags) {
		h.t.Errorf("%s did not fail", what)
	}
}

func TestResourceAPIErrors(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_tenant"
	typ := h.schemas.ResourceSchemas[name].ValueType()
	cfg := map[string]tftypes.Value{"name": str("x")}

	// create
	fake.failOn("POST", "tenancy/tenants", 500)
	diags, _ := h.tryApply(name, nullObject(), cfg)
	h.wantError("a create that NetBox rejects", diags)
	fake.clearFailures()

	// the object is created, but cannot be read back
	fake.failOn("GET", "tenancy/tenants/", 500)
	diags, _ = h.tryApply(name, nullObject(), cfg)
	h.wantError("a create whose read back fails", diags)
	fake.clearFailures()

	state := h.apply(name, nullObject(), cfg)

	// read
	fake.failOn("GET", "tenancy/tenants/", 500)
	h.wantError("a read that NetBox rejects", h.read(name, state).Diagnostics)

	// update, with the read back failing
	diags, _ = h.tryApply(name, state, map[string]tftypes.Value{"name": str("y")})
	h.wantError("an update whose read back fails", diags)
	fake.clearFailures()

	// update
	fake.failOn("PATCH", "tenancy/tenants/", 500)
	diags, _ = h.tryApply(name, state, map[string]tftypes.Value{"name": str("y")})
	h.wantError("an update that NetBox rejects", diags)
	fake.clearFailures()

	// delete
	fake.failOn("DELETE", "tenancy/tenants/", 500)
	h.wantError("a delete that NetBox rejects", h.destroy(name, state))
	fake.clearFailures()
	fake.failOn("DELETE", "tenancy/tenants/", 404)
	h.noErrors("a delete of an object that is gone", h.destroy(name, state))
	fake.clearFailures()

	// an ID in the state that is not a number
	bad := h.object(typ, map[string]tftypes.Value{"id": str("abc"), "name": str("x")})
	h.wantError("a read with a bad ID", h.read(name, bad).Diagnostics)
	h.wantError("a delete with a bad ID", h.destroy(name, bad))
	diags, _ = h.tryApply(name, bad, map[string]tftypes.Value{"name": str("y")})
	h.wantError("an update with a bad ID", diags)
}

func TestAssignAPIErrors(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_device_primary_ip"
	cfg := func(ip int64) map[string]tftypes.Value {
		return map[string]tftypes.Value{"device_id": num(5), "ip_address_id": num(ip), "ip_address_version": num(4)}
	}
	fake.put("dcim/devices", map[string]any{"id": float64(5), "name": "d"})

	fake.failOn("PATCH", "dcim/devices", 500)
	diags, _ := h.tryApply(name, nullObject(), cfg(10))
	h.wantError("an assignment that NetBox rejects", diags)
	fake.clearFailures()

	state := h.apply(name, nullObject(), cfg(10))
	fake.failOn("PATCH", "dcim/devices", 500)
	diags, _ = h.tryApply(name, state, cfg(11))
	h.wantError("a changed assignment that NetBox rejects", diags)
	fake.clearFailures()

	fake.failOn("GET", "dcim/devices", 500)
	h.wantError("a read that NetBox rejects", h.read(name, state).Diagnostics)
	h.wantError("a delete whose read fails", h.destroy(name, state))
	fake.clearFailures()

	// the assignment is still ours, but clearing it fails
	fake.get("dcim/devices", 5)["primary_ip4"] = map[string]any{"id": float64(10)}
	fake.failOn("PATCH", "dcim/devices", 500)
	h.wantError("a delete that NetBox rejects", h.destroy(name, state))
	fake.clearFailures()
	fake.failOn("PATCH", "dcim/devices", 404)
	h.noErrors("a delete of a device that is gone", h.destroy(name, state))
	fake.clearFailures()

	// an import of a device that only has an IPv6 address finds the version
	fake.put("dcim/devices", map[string]any{"id": float64(6), "name": "d6", "primary_ip6": map[string]any{"id": float64(20)}})
	imp, err := h.server.ImportResourceState(context.Background(), &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: "6"})
	if err != nil {
		t.Fatal(err)
	}
	rr, err := h.server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: imp.ImportedResources[0].State})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("read of the import", rr.Diagnostics)
	got := h.decode(h.schemas.ResourceSchemas[name].ValueType(), rr.NewState)
	if got["ip_address_version"].String() != num(6).String() {
		t.Errorf("version = %s, want 6", got["ip_address_version"])
	}
}

func TestDeviceTagAPIErrors(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_device_tag"
	typ := h.schemas.ResourceSchemas[name].ValueType()
	cfg := map[string]tftypes.Value{"device_id": num(7), "tag": str("blue")}

	fake.put("dcim/devices", map[string]any{"id": float64(7), "name": "d", "tags": []any{}})
	fake.put("extras/tags", map[string]any{"id": float64(1), "name": "blue", "slug": "blue"})

	fake.failOn("GET", "dcim/devices", 500)
	diags, _ := h.tryApply(name, nullObject(), cfg)
	h.wantError("a create that cannot read the device", diags)
	fake.clearFailures()

	fake.failOn("PATCH", "dcim/devices", 500)
	diags, _ = h.tryApply(name, nullObject(), cfg)
	h.wantError("a create that cannot update the device", diags)
	fake.clearFailures()

	state := h.apply(name, nullObject(), cfg)
	// creating it again when the device already has the tag is fine
	h.apply(name, nullObject(), cfg)

	fake.failOn("GET", "dcim/devices", 500)
	h.wantError("a read that NetBox rejects", h.read(name, state).Diagnostics)
	h.wantError("a delete whose read fails", h.destroy(name, state))
	fake.clearFailures()

	fake.failOn("GET", "dcim/devices", 404)
	if rr := h.read(name, state); h.hasError(rr.Diagnostics) {
		t.Error("a read of a device that is gone failed")
	}
	h.noErrors("a delete of a device that is gone", h.destroy(name, state))
	fake.clearFailures()

	fake.failOn("PATCH", "dcim/devices", 500)
	h.wantError("a delete that cannot update the device", h.destroy(name, state))
	fake.clearFailures()

	bad := h.object(typ, map[string]tftypes.Value{"id": str("nonsense"), "device_id": num(7), "tag": str("blue")})
	h.wantError("a read with a bad ID", h.read(name, bad).Diagnostics)
	h.wantError("a delete with a bad ID", h.destroy(name, bad))
}

func TestDataSourceAPIErrors(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()
	fake.put("tenancy/tenants", map[string]any{"id": float64(1), "name": "a", "slug": "a"})

	read := func(name string, set map[string]tftypes.Value) []*tfprotov6.Diagnostic {
		typ := h.schemas.DataSourceSchemas[name].ValueType()
		resp, err := h.server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{TypeName: name, Config: h.dynamic(h.object(typ, set))})
		if err != nil {
			t.Fatal(err)
		}
		return resp.Diagnostics
	}
	filter := func(name, value string) tftypes.Value {
		ft := h.schemas.DataSourceSchemas["netbox_tenants"].ValueType().(tftypes.Object).AttributeTypes["filter"]
		var elem tftypes.Type
		switch x := ft.(type) {
		case tftypes.Set:
			elem = x.ElementType
		case tftypes.List:
			elem = x.ElementType
		}
		return tftypes.NewValue(ft, []tftypes.Value{tftypes.NewValue(elem, map[string]tftypes.Value{"name": str(name), "value": str(value)})})
	}

	h.wantError("a filter that is not supported", read("netbox_tenants", map[string]tftypes.Value{"filter": filter("bogus", "x")}))
	h.wantError("an invalid regular expression", read("netbox_clusters", map[string]tftypes.Value{"name_regex": str("(")}))

	fake.failOn("GET", "tenancy/tenants", 500)
	h.wantError("a list that NetBox rejects", read("netbox_tenants", nil))
	h.wantError("a lookup that NetBox rejects", read("netbox_tenant", map[string]tftypes.Value{"name": str("a")}))
	fake.clearFailures()
}

func TestCustomDataSourceAPIErrors(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()
	fake.failOn("GET", "/api/ipam/", 500)
	fake.failOn("GET", "/api/dcim/", 500)
	fake.failOn("GET", "/api/extras/", 500)
	fake.failOn("POST", "/api/dcim/", 500)

	for _, tt := range []struct {
		name string
		set  map[string]tftypes.Value
	}{
		{"netbox_available_prefix", map[string]tftypes.Value{"prefix_id": num(1)}},
		{"netbox_device_render_config", map[string]tftypes.Value{"device_id": num(1)}},
		{"netbox_ip_address", map[string]tftypes.Value{"id": num(1)}},
		{"netbox_notification_group", map[string]tftypes.Value{"name": str("ops")}},
		{"netbox_rack_type", map[string]tftypes.Value{"model": str("m")}},
	} {
		typ := h.schemas.DataSourceSchemas[tt.name].ValueType()
		resp, err := h.server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{TypeName: tt.name, Config: h.dynamic(h.object(typ, tt.set))})
		if err != nil {
			t.Fatal(err)
		}
		h.wantError(tt.name+" with a failing NetBox", resp.Diagnostics)
	}
}

func TestConfigureVariants(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, server: server}
	h.schemas, _ = server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	typ := h.schemas.Provider.ValueType()
	configure := func(set map[string]tftypes.Value) []*tfprotov6.Diagnostic {
		set["server_url"] = str(srv.URL)
		set["api_token"] = str("token")
		resp, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: h.dynamic(h.object(typ, set))})
		if err != nil {
			t.Fatal(err)
		}
		return resp.Diagnostics
	}

	headers := tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"X-Test": str("1")})
	h.noErrors("configure with headers", configure(map[string]tftypes.Value{"headers": headers}))

	t.Setenv("NETBOX_HEADERS", `{"X-Env":"2"}`)
	h.noErrors("configure with headers from the environment", configure(map[string]tftypes.Value{}))
	t.Setenv("NETBOX_HEADERS", `not json`)
	h.wantError("headers that are not JSON", configure(map[string]tftypes.Value{}))
	t.Setenv("NETBOX_HEADERS", "")

	// a NetBox that rejects the token fails the configuration
	fake.failOn("GET", "status", 403)
	h.wantError("a NetBox that rejects the token", configure(map[string]tftypes.Value{}))
}

// With a provider that is not configured yet, every kind of resource and data source reports it.
func TestUnconfiguredProviderForEveryKind(t *testing.T) {
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, server: server}
	h.schemas, _ = server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	cfg := h.object(h.schemas.Provider.ValueType(), map[string]tftypes.Value{"server_url": tftypes.NewValue(tftypes.String, tftypes.UnknownValue)})
	if _, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: h.dynamic(cfg)}); err != nil {
		t.Fatal(err)
	}

	for name, set := range map[string]map[string]tftypes.Value{
		"netbox_tenant":            {"name": str("x")},
		"netbox_device_primary_ip": {"device_id": num(1), "ip_address_id": num(2), "ip_address_version": num(4)},
		"netbox_device_tag":        {"device_id": num(1), "tag": str("t")},
	} {
		typ := h.schemas.ResourceSchemas[name].ValueType()
		diags, _ := h.tryApply(name, nullObject(), set)
		h.wantError(name+" create", diags)
		state := h.object(typ, set)
		state = h.object(typ, mergeID(state, h, typ))
		h.wantError(name+" read", h.read(name, state).Diagnostics)
		h.wantError(name+" delete", h.destroy(name, state))
		if name != "netbox_device_tag" {
			changed := map[string]tftypes.Value{}
			for k, v := range set {
				changed[k] = v
			}
			if name == "netbox_tenant" {
				changed["name"] = str("y")
			} else {
				changed["ip_address_id"] = num(3)
			}
			diags, _ = h.tryApply(name, state, changed)
			h.wantError(name+" update", diags)
		}
	}

	typ := h.schemas.DataSourceSchemas["netbox_tenant"].ValueType()
	resp, err := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "netbox_tenant", Config: h.dynamic(h.object(typ, map[string]tftypes.Value{"name": str("x")}))})
	if err != nil {
		t.Fatal(err)
	}
	h.wantError("data source read", resp.Diagnostics)
}

// mergeID returns the attributes of a state value with an id added.
func mergeID(state tftypes.Value, h *harness, typ tftypes.Type) map[string]tftypes.Value {
	var m map[string]tftypes.Value
	_ = state.As(&m)
	m["id"] = str("1")
	return m
}

func TestUnknownTagFailsTheRequest(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_tenant"

	tags := setOf(tftypes.String, str("nonexistent"))
	diags, _ := h.tryApply(name, nullObject(), map[string]tftypes.Value{"name": str("x"), "tags": tags})
	h.wantError("a create with an unknown tag", diags)

	state := h.apply(name, nullObject(), map[string]tftypes.Value{"name": str("x")})
	diags, _ = h.tryApply(name, state, map[string]tftypes.Value{"name": str("x"), "tags": tags})
	h.wantError("an update with an unknown tag", diags)

	// the same while NetBox allocates the address, which is updated after the allocation
	diags, _ = h.tryApply("netbox_available_ip_address", nullObject(), map[string]tftypes.Value{"prefix_id": num(1), "tags": tags})
	h.wantError("an allocation with an unknown tag", diags)
}

func TestHooksThatFailFailTheRequest(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_device"
	fake.put("dcim/virtual-chassis", map[string]any{"id": float64(1), "name": "vc"})
	set := func(master bool) map[string]tftypes.Value {
		return map[string]tftypes.Value{
			"name": str("sw"), "device_type_id": num(1), "role_id": num(1), "site_id": num(1),
			"virtual_chassis_id": num(1), "virtual_chassis_master": tftypes.NewValue(tftypes.Bool, master),
		}
	}

	// the hook after the write cannot make the device the master
	fake.failOn("PATCH", "dcim/virtual-chassis", 500)
	diags, _ := h.tryApply(name, nullObject(), set(true))
	h.wantError("a create whose master cannot be set", diags)
	fake.clearFailures()

	state := h.apply(name, nullObject(), set(false))
	fake.failOn("PATCH", "dcim/virtual-chassis", 500)
	diags, _ = h.tryApply(name, state, set(true))
	h.wantError("an update whose master cannot be set", diags)
	fake.clearFailures()

	// the hook before the delete cannot release the master
	state = h.apply(name, state, set(true))
	id := int(numberAttr(t, state, "id"))
	fake.get("dcim/devices", id)["virtual_chassis"] = map[string]any{"id": float64(1), "master": map[string]any{"id": float64(id)}}
	state = h.readState(name, state)
	fake.failOn("PATCH", "dcim/virtual-chassis", 500)
	h.wantError("a delete whose master cannot be released", h.destroy(name, state))
}

func TestResponseWithoutAnID(t *testing.T) {
	noID := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/status/":
			_, _ = w.Write([]byte(`{"netbox-version":"4.4.0"}`))
		case strings.HasSuffix(r.URL.Path, "/available-ips/"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`[{"address":"10.0.0.1/24"}]`))
		default:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"name":"x"}`))
		}
	}))
	defer noID.Close()
	h := newHarness(t, noID.URL)

	diags, _ := h.tryApply("netbox_tenant", nullObject(), map[string]tftypes.Value{"name": str("x")})
	h.wantError("a create answered without an id", diags)
	diags, _ = h.tryApply("netbox_available_ip_address", nullObject(), map[string]tftypes.Value{"prefix_id": num(1)})
	h.wantError("an allocation answered without an id", diags)
}

func TestAssignReadsTheObjectBeforeKeepingVLANs(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_device_interface_primary_mac_address"
	fake.put("dcim/interfaces", map[string]any{"id": float64(4), "name": "eth0"})

	cfg := map[string]tftypes.Value{"interface_id": num(4), "mac_address_id": num(9)}
	fake.failOn("GET", "dcim/interfaces", 500)
	diags, _ := h.tryApply(name, nullObject(), cfg)
	h.wantError("an assignment that cannot read the interface", diags)
	fake.clearFailures()

	state := h.apply(name, nullObject(), cfg)
	fake.failOn("GET", "dcim/interfaces", 404)
	h.noErrors("a delete of an interface that is gone", h.destroy(name, state))
}

func TestVirtualDiskDataSourceOptions(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	fake.put("virtualization/virtual-disks", map[string]any{"id": float64(1), "name": "disk0", "size": float64(10)})

	typ := h.schemas.DataSourceSchemas["netbox_virtual_disk"].ValueType()
	order := tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{str("name"), str("-id")})
	resp, err := h.server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{
		TypeName: "netbox_virtual_disk",
		Config:   h.dynamic(h.object(typ, map[string]tftypes.Value{"name_regex": str("^disk"), "ordering": order})),
	})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("read", resp.Diagnostics)
}
