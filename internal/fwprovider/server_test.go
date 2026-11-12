package fwprovider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// fakeNetBox keeps objects in memory and answers the REST calls the provider makes.
type fakeNetBox struct {
	mu      sync.Mutex
	next    int
	objects map[string]map[int]map[string]any // collection path -> id -> object
	calls   []string
}

func newFakeNetBox() (*fakeNetBox, *httptest.Server) {
	f := &fakeNetBox{next: 1, objects: map[string]map[int]map[string]any{}}
	return f, httptest.NewServer(f)
}

func (f *fakeNetBox) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, r.Method+" "+r.URL.Path)
	w.Header().Set("Content-Type", "application/json")

	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/"), "/")
	if rest == "status" {
		_ = json.NewEncoder(w).Encode(map[string]any{"netbox-version": "4.4.0"})
		return
	}
	parts := strings.Split(rest, "/")
	id := 0
	collection := rest
	if n, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
		id = n
		collection = strings.Join(parts[:len(parts)-1], "/")
	}
	if f.objects[collection] == nil {
		f.objects[collection] = map[int]map[string]any{}
	}
	objs := f.objects[collection]

	if r.Method == http.MethodPost && (strings.HasSuffix(collection, "/available-prefixes") || strings.HasSuffix(collection, "/available-vlans")) {
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		target, obj := "ipam/prefixes", map[string]any{"id": float64(f.next)}
		if strings.HasSuffix(collection, "/available-vlans") {
			target = "ipam/vlans"
			obj["vid"] = float64(100 + f.next)
		} else {
			obj["prefix"] = "10.1." + strconv.Itoa(f.next) + ".0/24"
		}
		for k, v := range in {
			if k != "prefix_length" {
				obj[k] = v
			}
		}
		if f.objects[target] == nil {
			f.objects[target] = map[int]map[string]any{}
		}
		f.objects[target][f.next] = obj
		f.next++
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(obj)
		return
	}
	if r.Method == http.MethodPost && strings.HasSuffix(collection, "/available-ips") {
		// allocate the next free address
		var entries []map[string]any
		_ = json.NewDecoder(r.Body).Decode(&entries)
		if f.objects["ipam/ip-addresses"] == nil {
			f.objects["ipam/ip-addresses"] = map[int]map[string]any{}
		}
		var created []any
		for range entries {
			obj := map[string]any{"id": float64(f.next), "address": "10.0.0." + strconv.Itoa(f.next) + "/24"}
			f.objects["ipam/ip-addresses"][f.next] = obj
			f.next++
			created = append(created, obj)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
		return
	}

	var body map[string]any
	if r.Body != nil && (r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodPut) {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	switch {
	case r.Method == http.MethodPost:
		obj := map[string]any{"id": float64(f.next)}
		for k, v := range body {
			obj[k] = v
		}
		objs[f.next] = obj
		f.next++
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(obj)
	case r.Method == http.MethodGet && id == 0:
		var results []any
		for _, o := range objs {
			match := true
			for k, v := range r.URL.Query() {
				switch k {
				case "limit", "offset", "brief", "ordering", "q", "contains", "vrf_id":
					continue
				}
				if s, ok := o[k].(string); !ok || s != v[0] {
					match = false
				}
			}
			if match {
				results = append(results, o)
			}
		}
		if results == nil {
			results = []any{}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"count": len(results), "results": results})
	default:
		obj, ok := objs[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"Not found."}`))
			return
		}
		switch r.Method {
		case http.MethodPatch, http.MethodPut:
			for k, v := range body {
				obj[k] = v
			}
		case http.MethodDelete:
			delete(objs, id)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(obj)
	}
}

// harness drives the provider through the plugin protocol, like Terraform does.
type harness struct {
	t       *testing.T
	server  tfprotov6.ProviderServer
	schemas *tfprotov6.GetProviderSchemaResponse
}

func newHarness(t *testing.T, url string) *harness {
	t.Helper()
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, server: server}
	h.schemas, err = server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range h.schemas.Diagnostics {
		t.Fatalf("schema: %s: %s", d.Summary, d.Detail)
	}

	cfg := h.object(h.schemas.Provider.ValueType(), map[string]tftypes.Value{
		"server_url": tftypes.NewValue(tftypes.String, url),
		"api_token":  tftypes.NewValue(tftypes.String, "token"),
	})
	resp, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("configure", resp.Diagnostics)
	return h
}

func (h *harness) noErrors(what string, diags []*tfprotov6.Diagnostic) {
	h.t.Helper()
	for _, d := range diags {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			h.t.Fatalf("%s: %s: %s", what, d.Summary, d.Detail)
		}
	}
}

func (h *harness) hasError(diags []*tfprotov6.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

// object builds an object value of the given type in which attributes that are not set are null.
func (h *harness) object(typ tftypes.Type, set map[string]tftypes.Value) tftypes.Value {
	h.t.Helper()
	ot := typ.(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for name, at := range ot.AttributeTypes {
		if v, ok := set[name]; ok {
			vals[name] = v
		} else {
			vals[name] = tftypes.NewValue(at, nil)
		}
	}
	return tftypes.NewValue(typ, vals)
}

func (h *harness) dynamic(v tftypes.Value) *tfprotov6.DynamicValue {
	h.t.Helper()
	d, err := tfprotov6.NewDynamicValue(v.Type(), v)
	if err != nil {
		h.t.Fatal(err)
	}
	return &d
}

func (h *harness) decode(typ tftypes.Type, d *tfprotov6.DynamicValue) map[string]tftypes.Value {
	h.t.Helper()
	v, err := d.Unmarshal(typ)
	if err != nil {
		h.t.Fatal(err)
	}
	var out map[string]tftypes.Value
	if err := v.As(&out); err != nil {
		h.t.Fatal(err)
	}
	return out
}

// apply plans and applies the change from prior to config, and returns the new state.
func (h *harness) apply(typeName string, prior tftypes.Value, set map[string]tftypes.Value) tftypes.Value {
	h.t.Helper()
	ctx := context.Background()
	typ := h.schemas.ResourceSchemas[typeName].ValueType()
	cfg := h.object(typ, set)
	if prior.IsNull() {
		prior = tftypes.NewValue(typ, nil)
	}

	vr, err := h.server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{TypeName: typeName, Config: h.dynamic(cfg)})
	if err != nil {
		h.t.Fatal(err)
	}
	h.noErrors("validate", vr.Diagnostics)

	// Terraform proposes the configuration with the prior values of computed attributes.
	proposed := h.proposed(typeName, typ, prior, cfg)
	plan, err := h.server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName:         typeName,
		PriorState:       h.dynamic(prior),
		ProposedNewState: h.dynamic(proposed),
		Config:           h.dynamic(cfg),
	})
	if err != nil {
		h.t.Fatal(err)
	}
	h.noErrors("plan", plan.Diagnostics)

	applied, err := h.server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{
		TypeName:       typeName,
		PriorState:     h.dynamic(prior),
		PlannedState:   plan.PlannedState,
		Config:         h.dynamic(cfg),
		PlannedPrivate: plan.PlannedPrivate,
	})
	if err != nil {
		h.t.Fatal(err)
	}
	h.noErrors("apply", applied.Diagnostics)
	v, err := applied.NewState.Unmarshal(typ)
	if err != nil {
		h.t.Fatal(err)
	}
	return v
}

func (h *harness) proposed(typeName string, typ tftypes.Type, prior, cfg tftypes.Value) tftypes.Value {
	computed := map[string]bool{}
	for _, a := range h.schemas.ResourceSchemas[typeName].Block.Attributes {
		computed[a.Name] = a.Computed
	}
	var p, c map[string]tftypes.Value
	_ = cfg.As(&c)
	if prior.IsNull() {
		return cfg
	}
	_ = prior.As(&p)
	out := map[string]tftypes.Value{}
	for k, v := range c {
		if v.IsNull() && computed[k] && p[k].IsKnown() {
			out[k] = p[k]
		} else {
			out[k] = v
		}
	}
	return tftypes.NewValue(typ, out)
}

func (h *harness) read(typeName string, state tftypes.Value) *tfprotov6.ReadResourceResponse {
	h.t.Helper()
	resp, err := h.server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: typeName, CurrentState: h.dynamic(state)})
	if err != nil {
		h.t.Fatal(err)
	}
	return resp
}

func (h *harness) destroy(typeName string, state tftypes.Value) []*tfprotov6.Diagnostic {
	h.t.Helper()
	typ := h.schemas.ResourceSchemas[typeName].ValueType()
	null := tftypes.NewValue(typ, nil)
	resp, err := h.server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{
		TypeName:     typeName,
		PriorState:   h.dynamic(state),
		PlannedState: h.dynamic(null),
		Config:       h.dynamic(null),
	})
	if err != nil {
		h.t.Fatal(err)
	}
	return resp.Diagnostics
}

func str(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }

func attrOf(t *testing.T, state tftypes.Value, name string) tftypes.Value {
	t.Helper()
	var m map[string]tftypes.Value
	if err := state.As(&m); err != nil {
		t.Fatal(err)
	}
	return m[name]
}

func stringAttr(t *testing.T, state tftypes.Value, name string) string {
	t.Helper()
	v := attrOf(t, state, name)
	var s string
	if err := v.As(&s); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return s
}

func TestResourceLifecycle(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()
	const name = "netbox_tenant"

	// create
	state := h.apply(name, tftypes.NewValue(tftypes.Object{}, nil), map[string]tftypes.Value{
		"name":        str("acme"),
		"description": str("first"),
	})
	id := stringAttr(t, state, "id")
	if id == "" {
		t.Fatal("no id in the state")
	}
	if got := stringAttr(t, state, "slug"); got != "acme" {
		t.Errorf("slug = %q, want acme", got)
	}

	// refresh keeps the state
	rr := h.read(name, state)
	h.noErrors("read", rr.Diagnostics)
	back := h.decode(h.schemas.ResourceSchemas[name].ValueType(), rr.NewState)
	if back["name"].String() != attrOf(t, state, "name").String() {
		t.Errorf("name changed on refresh: %s", back["name"])
	}

	// update
	state = h.apply(name, state, map[string]tftypes.Value{
		"name":        str("acme"),
		"description": str("second"),
	})
	if got := stringAttr(t, state, "description"); got != "second" {
		t.Errorf("description = %q, want second", got)
	}

	// import
	imp, err := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: id})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("import", imp.Diagnostics)
	if len(imp.ImportedResources) != 1 {
		t.Fatalf("imported %d resources", len(imp.ImportedResources))
	}
	if _, err := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: "abc"}); err != nil {
		t.Fatal(err)
	}

	// delete, then a refresh finds nothing
	h.noErrors("destroy", h.destroy(name, state))
	rr = h.read(name, state)
	h.noErrors("read after delete", rr.Diagnostics)
	if rr.NewState != nil {
		v, _ := rr.NewState.Unmarshal(h.schemas.ResourceSchemas[name].ValueType())
		if !v.IsNull() {
			t.Error("the deleted object is still in the state")
		}
	}

	if len(fake.calls) == 0 {
		t.Error("the fake NetBox was never called")
	}
}

func TestUnconfiguredProvider(t *testing.T) {
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, server: server}
	h.schemas, _ = server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})

	// Unknown settings leave the provider unconfigured. A resource must say so instead of crashing.
	cfg := h.object(h.schemas.Provider.ValueType(), map[string]tftypes.Value{
		"server_url": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	})
	resp, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: h.dynamic(cfg)})
	if err != nil || h.hasError(resp.Diagnostics) {
		t.Fatalf("configure: %v %v", err, resp.Diagnostics)
	}

	typ := h.schemas.ResourceSchemas["netbox_tenant"].ValueType()
	state := h.object(typ, map[string]tftypes.Value{"id": str("1"), "name": str("x")})
	rr := h.read("netbox_tenant", state)
	if !h.hasError(rr.Diagnostics) {
		t.Error("read without a configured provider did not fail")
	}
	if !h.hasError(h.destroy("netbox_tenant", state)) {
		t.Error("destroy without a configured provider did not fail")
	}
}

func TestMissingSettings(t *testing.T) {
	t.Setenv("NETBOX_SERVER_URL", "")
	t.Setenv("NETBOX_API_TOKEN", "")
	server, _ := providerserver.NewProtocol6WithError(New("test")())()
	h := &harness{t: t, server: server}
	h.schemas, _ = server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	cfg := h.object(h.schemas.Provider.ValueType(), nil)
	resp, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	if !h.hasError(resp.Diagnostics) {
		t.Error("missing server_url and api_token were accepted")
	}
}

func TestDataSourceRead(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()

	h.apply("netbox_tenant", tftypes.NewValue(tftypes.Object{}, nil), map[string]tftypes.Value{"name": str("looked-up")})

	typ := h.schemas.DataSourceSchemas["netbox_tenant"].ValueType()
	cfg := h.object(typ, map[string]tftypes.Value{"name": str("looked-up")})
	resp, err := h.server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{TypeName: "netbox_tenant", Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("read data source", resp.Diagnostics)
	out := h.decode(typ, resp.State)
	var id string
	if err := out["id"].As(&id); err != nil || id == "" {
		t.Errorf("data source has no id: %v", out["id"])
	}

	// nothing found is an error
	cfg = h.object(typ, map[string]tftypes.Value{"name": str("missing")})
	resp, err = h.server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{TypeName: "netbox_tenant", Config: h.dynamic(cfg)})
	if err != nil {
		t.Fatal(err)
	}
	if !h.hasError(resp.Diagnostics) {
		t.Error("a data source that matches nothing did not fail")
	}
}

// Every resource and data source must produce a valid schema, and every resource must accept an
// empty configuration in validation without crashing.
func TestAllSchemas(t *testing.T) {
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, server: server}
	h.schemas, err = server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("schema", h.schemas.Diagnostics)
	if len(h.schemas.ResourceSchemas) < 100 || len(h.schemas.DataSourceSchemas) < 50 {
		t.Fatalf("only %d resources and %d data sources", len(h.schemas.ResourceSchemas), len(h.schemas.DataSourceSchemas))
	}
	for name, s := range h.schemas.ResourceSchemas {
		cfg := h.object(s.ValueType(), nil)
		if _, err := server.ValidateResourceConfig(context.Background(), &tfprotov6.ValidateResourceConfigRequest{TypeName: name, Config: h.dynamic(cfg)}); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	for name, s := range h.schemas.DataSourceSchemas {
		cfg := h.object(s.ValueType(), nil)
		if _, err := server.ValidateDataResourceConfig(context.Background(), &tfprotov6.ValidateDataResourceConfigRequest{TypeName: name, Config: h.dynamic(cfg)}); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// If any provider setting is not known yet, the provider stays unconfigured.
func TestUnknownSettingsLeaveProviderUnconfigured(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	typ := h.schemas.Provider.ValueType()
	cfg := h.object(typ, map[string]tftypes.Value{
		"server_url":   str(srv.URL),
		"api_token":    str("t"),
		"default_tags": tftypes.NewValue(typ.(tftypes.Object).AttributeTypes["default_tags"], tftypes.UnknownValue),
	})
	resp, err := h.server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: h.dynamic(cfg)})
	if err != nil || h.hasError(resp.Diagnostics) {
		t.Fatalf("configure: %v %v", err, resp.Diagnostics)
	}
	rtyp := h.schemas.ResourceSchemas["netbox_tenant"].ValueType()
	state := h.object(rtyp, map[string]tftypes.Value{"id": str("1"), "name": str("x")})
	if !h.hasError(h.read("netbox_tenant", state).Diagnostics) {
		t.Error("the provider was configured although default_tags is unknown")
	}
}
