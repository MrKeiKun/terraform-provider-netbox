package fwprovider

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func specOf(t *testing.T, name string) *fwresource.Spec {
	t.Helper()
	for _, s := range standardResources {
		if s.TypeName == name {
			return s
		}
	}
	t.Fatalf("no spec for %s", name)
	return nil
}

// NetBox never returns a password, so a user needs one when it is created and not afterwards.
func TestUserNeedsAPasswordOnlyToCreate(t *testing.T) {
	s := specOf(t, "netbox_user")
	plan := fwresource.Values{"password": types.StringNull()}
	if err := s.ModifyBody(context.Background(), plan, nil, map[string]any{}); err == nil {
		t.Error("a user without a password was accepted on create")
	}
	if err := s.ModifyBody(context.Background(), plan, fwresource.Values{}, map[string]any{}); err != nil {
		t.Errorf("an update without a password failed: %v", err)
	}
	plan["password"] = types.StringValue("secret")
	if err := s.ModifyBody(context.Background(), plan, nil, map[string]any{}); err != nil {
		t.Errorf("a user with a password failed: %v", err)
	}
}

// A rack with a rack type gets its dimensions from the type, so the state does not hold them and a
// configuration written from the state does not conflict with rack_type_id.
func TestRackWithRackTypeDoesNotKeepDerivedDimensions(t *testing.T) {
	s := specOf(t, "netbox_rack")
	out := fwresource.Values{"form_factor": types.StringValue("4-post-cabinet"), "width": types.Int64Value(19), "u_height": types.Int64Value(42)}
	if err := s.ModifyState(context.Background(), map[string]any{"rack_type": map[string]any{"id": float64(1)}}, nil, out); err != nil {
		t.Fatal(err)
	}
	if !out["form_factor"].IsNull() || !out["width"].IsNull() || !out["u_height"].IsNull() {
		t.Errorf("derived dimensions stayed in the state: %v", out)
	}

	out = fwresource.Values{"form_factor": types.StringValue("4-post-cabinet"), "width": types.Int64Value(19), "u_height": types.Int64Value(42)}
	_ = s.ModifyState(context.Background(), map[string]any{"rack_type": nil}, nil, out)
	if out["u_height"].IsNull() {
		t.Error("a rack without a rack type lost its height")
	}
}

// The values of these attributes come back from NetBox, so leaving them out of the configuration
// must not clear them.
func TestSecretsAreKeptWhenNotConfigured(t *testing.T) {
	for typeName, attr := range map[string]string{
		"netbox_ike_policy":    "preshared_key",
		"netbox_fhrp_group":    "auth_key",
		"netbox_webhook":       "additional_headers",
		"netbox_wireless_lan":  "auth_psk",
		"netbox_wireless_link": "auth_psk",
		"netbox_data_source":   "parameters",
	} {
		found := false
		for _, a := range specOf(t, typeName).Attrs {
			if a.Name == attr {
				found = true
				if !a.Computed || !a.Sensitive {
					t.Errorf("%s.%s: computed=%v sensitive=%v, want both", typeName, attr, a.Computed, a.Sensitive)
				}
			}
		}
		if !found {
			t.Errorf("%s has no %s", typeName, attr)
		}
	}
}

// NetBox derives the data source, the data path and the automatic sync from the data file, so the plan
// has to predict that they are cleared when the data file is taken away.
func TestDataFilePlanClearsWhatNetBoxDerives(t *testing.T) {
	plan := fwresource.Values{
		"data_source_id":    types.Int64Value(5),
		"data_path":         types.StringValue("t.j2"),
		"auto_sync_enabled": types.BoolValue(true),
	}
	prior := fwresource.Values{"data_file_id": types.Int64Value(1)}

	got := dataFilePlan(fwresource.Values{"data_file_id": types.Int64Null()}, plan, prior)
	if got["data_source_id"].String() != "0" || got["data_path"].String() != `""` || got["auto_sync_enabled"].String() != "false" {
		t.Errorf("unlinking did not clear the derived values: %v", got)
	}
	if got := dataFilePlan(fwresource.Values{"data_file_id": types.Int64Value(1)}, plan, prior); got != nil {
		t.Errorf("a linked object was changed: %v", got)
	}
	if got := dataFilePlan(fwresource.Values{"data_file_id": types.Int64Null()}, plan, nil); got != nil {
		t.Errorf("a new object was changed: %v", got)
	}
	// a data source without a data file (older NetBox versions keep one) is left alone
	if got := dataFilePlan(fwresource.Values{"data_file_id": types.Int64Null()}, plan, fwresource.Values{"data_file_id": types.Int64Value(0)}); got != nil {
		t.Errorf("an object that never had a data file was changed: %v", got)
	}
}

// An old attribute name and its new name stay equal in the plan: the one that is left out takes the
// value of the other.
func TestAliasPlanFollowsTheNameThatIsWritten(t *testing.T) {
	set := portSet("dcim.site")
	null := types.SetNull(types.StringType)
	plan := fwresource.Values{"content_types": set, "object_types": types.SetUnknown(types.StringType)}
	got := aliasPlan("content_types", "object_types")(fwresource.Values{"content_types": set, "object_types": null}, plan, nil)
	if got["object_types"] == nil || !got["object_types"].Equal(set) {
		t.Errorf("the new name did not follow the old one: %v", got)
	}
	plan = fwresource.Values{"content_types": types.SetUnknown(types.StringType), "object_types": set}
	got = aliasPlan("content_types", "object_types")(fwresource.Values{"content_types": null, "object_types": set}, plan, nil)
	if got["content_types"] == nil || !got["content_types"].Equal(set) {
		t.Errorf("the old name did not follow the new one: %v", got)
	}
	if got := aliasPlan("content_types", "object_types")(fwresource.Values{"content_types": set, "object_types": set}, plan, nil); got != nil {
		t.Errorf("both written, nothing to follow: %v", got)
	}
}

func TestDataFileBodyLeavesOutANullDataFile(t *testing.T) {
	create := map[string]any{"data_file": nil, "name": "a"}
	if err := dataFileBody(context.Background(), nil, nil, create); err != nil {
		t.Fatal(err)
	}
	if _, ok := create["data_file"]; ok {
		t.Fatalf("a create must not send data_file: null, got %v", create)
	}

	hadFile := fwresource.Values{"data_file_id": types.Int64Value(7)}
	update := map[string]any{"data_file": nil}
	if err := dataFileBody(context.Background(), nil, hadFile, update); err != nil {
		t.Fatal(err)
	}
	if _, ok := update["data_file"]; !ok {
		t.Fatal("taking a data file away must still send null")
	}
}

func TestUserKeepsStaffWhenNetBoxHasNoIsStaff(t *testing.T) {
	s := specOf(t, "netbox_user")
	prior := fwresource.Values{"staff": types.BoolValue(true)}
	out := fwresource.Values{"staff": types.BoolValue(false)}
	if err := s.ModifyState(context.Background(), map[string]any{"username": "a"}, prior, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["staff"].(types.Bool); !v.ValueBool() {
		t.Error("staff = true was not kept when NetBox returns no is_staff")
	}
	out = fwresource.Values{"staff": types.BoolValue(false)}
	if err := s.ModifyState(context.Background(), map[string]any{"is_staff": false}, prior, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["staff"].(types.Bool); v.ValueBool() {
		t.Error("the value NetBox returns was overwritten")
	}
}

func TestDataFileBodyKeepsOtherValues(t *testing.T) {
	body := map[string]any{"data_file": float64(5)}
	if err := dataFileBody(context.Background(), nil, nil, body); err != nil {
		t.Fatal(err)
	}
	if body["data_file"] != float64(5) {
		t.Errorf("a data file was changed: %v", body)
	}
	empty := map[string]any{"name": "a"}
	if err := dataFileBody(context.Background(), nil, nil, empty); err != nil || len(empty) != 1 {
		t.Errorf("a body without data_file was changed: %v", empty)
	}
}

func TestDataFileStateKeepsDataSourceAtZeroWithoutOne(t *testing.T) {
	out := fwresource.Values{"data_source_id": types.Int64Value(3)}
	if err := dataFileState(context.Background(), map[string]any{"data_source": nil}, nil, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["data_source_id"].(types.Int64); v.ValueInt64() != 0 {
		t.Errorf("data_source_id = %v, want 0", out["data_source_id"])
	}
	out = fwresource.Values{"data_source_id": types.Int64Value(3)}
	if err := dataFileState(context.Background(), map[string]any{"data_source": map[string]any{"id": 3}}, nil, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["data_source_id"].(types.Int64); v.ValueInt64() != 3 {
		t.Errorf("a data source was dropped: %v", out["data_source_id"])
	}
}

// An update that links a data file sets the data source in NetBox, also when the object was created
// with the file and has no data source yet.
func TestDataFilePlanExpectsTheDataSourceAfterALinkingUpdate(t *testing.T) {
	updating := func(file, source int64) (fwresource.Values, fwresource.Values) {
		plan := fwresource.Values{
			"data_file_id":   types.Int64Value(file),
			"data_source_id": types.Int64Value(source),
			"last_updated":   types.StringUnknown(),
		}
		return plan, fwresource.Values{"data_file_id": types.Int64Value(file), "data_source_id": types.Int64Value(source)}
	}
	cfg := fwresource.Values{"data_file_id": types.Int64Value(1)}

	// linked, but never synced (data source 0): NetBox sets it
	plan, prior := updating(1, 0)
	if got := dataFilePlan(cfg, plan, prior); got["data_source_id"] == nil || !got["data_source_id"].IsUnknown() {
		t.Errorf("the data source was not marked unknown: %v", got)
	}
	// linked and synced already: nothing changes
	plan, prior = updating(1, 4)
	if got := dataFilePlan(cfg, plan, prior); got != nil {
		t.Errorf("a synced object was changed: %v", got)
	}
	// another file: the source can change
	plan, prior = updating(2, 4)
	prior["data_file_id"] = types.Int64Value(1)
	if got := dataFilePlan(cfg, plan, prior); got["data_source_id"] == nil || !got["data_source_id"].IsUnknown() {
		t.Errorf("a new file did not mark the data source unknown: %v", got)
	}
	// no update planned (last_updated is not unknown): nothing may change, or the plan would never settle
	plan, prior = updating(1, 0)
	plan["last_updated"] = types.StringValue("2026-01-01T00:00:00Z")
	if got := dataFilePlan(cfg, plan, prior); got != nil {
		t.Errorf("a plan without an update was changed: %v", got)
	}
	// no file: nothing to link
	plan, prior = updating(0, 0)
	if got := dataFilePlan(fwresource.Values{"data_file_id": types.Int64Value(0)}, plan, prior); got != nil {
		t.Errorf("an object without a file was changed: %v", got)
	}
}

func choiceList(values ...string) types.List {
	inner := types.ListType{ElemType: types.StringType}
	var elems []attr.Value
	for _, v := range values {
		l, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue(v), types.StringValue("label " + v)})
		elems = append(elems, l)
	}
	l, _ := types.ListValue(inner, elems)
	return l
}

// With order_alphabetically NetBox sorts the choices by their value, so a list in another order has to
// be refused before the apply.
func TestChoiceSetOrder(t *testing.T) {
	on, off := types.BoolValue(true), types.BoolValue(false)
	for name, tc := range map[string]struct {
		alpha   attr.Value
		choices attr.Value
		bad     bool
	}{
		"sorted":                   {on, choiceList("a", "b", "c"), false},
		"unsorted":                 {on, choiceList("b", "a"), true},
		"unsorted but not ordered": {off, choiceList("b", "a"), false},
		"capitals before letters":  {on, choiceList("A", "B", "a"), false},
		"letters before capitals":  {on, choiceList("a", "B"), true},
		"digits are text":          {on, choiceList("10", "2", "9"), false},
		"equal values keep order":  {on, choiceList("a", "a"), false},
		"alpha not set":            {types.BoolNull(), choiceList("b", "a"), false},
		"alpha unknown":            {types.BoolUnknown(), choiceList("b", "a"), false},
		"choices unknown":          {on, types.ListUnknown(types.ListType{ElemType: types.StringType}), false},
		"choices not set":          {on, types.ListNull(types.ListType{ElemType: types.StringType}), false},
		"single choice":            {on, choiceList("z"), false},
	} {
		d := choiceSetOrder(context.Background(), fwresource.Values{"order_alphabetically": tc.alpha, "extra_choices": tc.choices})
		if d.HasError() != tc.bad {
			t.Errorf("%s: error = %v, want %v (%v)", name, d.HasError(), tc.bad, d)
		}
	}
	d := choiceSetOrder(context.Background(), fwresource.Values{"order_alphabetically": on, "extra_choices": choiceList("c", "a", "b")})
	if len(d) != 1 || !strings.Contains(d[0].Detail(), "a, b, c") {
		t.Errorf("the message does not give the order: %v", d)
	}
}

// NetBox 4.4 does not return auto_sync_enabled, so the state keeps what was written instead of null.
func TestDataFileStateKeepsAutoSyncWhenNetBoxHasNone(t *testing.T) {
	out := fwresource.Values{"auto_sync_enabled": types.BoolNull(), "data_source_id": types.Int64Value(0)}
	prior := fwresource.Values{"auto_sync_enabled": types.BoolValue(false)}
	if err := dataFileState(context.Background(), map[string]any{"data_source": nil}, prior, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["auto_sync_enabled"].(types.Bool); v.IsNull() || v.ValueBool() {
		t.Errorf("auto_sync_enabled = %v, want false", out["auto_sync_enabled"])
	}
	// nothing written yet (unknown while the object is created): false
	out = fwresource.Values{"auto_sync_enabled": types.BoolNull()}
	if err := dataFileState(context.Background(), map[string]any{}, fwresource.Values{"auto_sync_enabled": types.BoolUnknown()}, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["auto_sync_enabled"].(types.Bool); v.IsNull() || v.ValueBool() {
		t.Errorf("an unknown value became %v, want false", out["auto_sync_enabled"])
	}
	// a NetBox that returns the field keeps what it says
	out = fwresource.Values{"auto_sync_enabled": types.BoolValue(true)}
	if err := dataFileState(context.Background(), map[string]any{"auto_sync_enabled": true}, nil, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["auto_sync_enabled"].(types.Bool); !v.ValueBool() {
		t.Errorf("the value of NetBox was overwritten: %v", out["auto_sync_enabled"])
	}
}

// NetBox 4.7.5 and later replace the content with the text of the data file. The error names what
// happened, and nothing is said when the text is the same.
func TestDataFileWrittenReportsReplacedContent(t *testing.T) {
	plan := func(file int64, text string) fwresource.Values {
		return fwresource.Values{"data_file_id": types.Int64Value(file), "template_code": types.StringValue(text)}
	}
	check := dataFileWritten("template_code", false)
	obj := map[string]any{"data_file": map[string]any{"id": 3.0}, "data_path": "t.j2", "template_code": "FILE\n"}

	if err := check(plan(3, "FILE\n"), obj); err != nil {
		t.Errorf("the same text was reported: %v", err)
	}
	err := check(plan(3, "MINE\n"), obj)
	if err == nil || !strings.Contains(err.Error(), "t.j2") || !strings.Contains(err.Error(), "remove data_file_id") {
		t.Errorf("a replaced text was not explained: %v", err)
	}
	if err := check(plan(0, "MINE\n"), obj); err != nil {
		t.Errorf("an object without a file was reported: %v", err)
	}
	if err := check(plan(3, "MINE\n"), map[string]any{"data_file": nil, "template_code": "other"}); err != nil {
		t.Errorf("an object that is not linked was reported: %v", err)
	}
	if err := check(fwresource.Values{"data_file_id": types.Int64Value(3), "template_code": types.StringNull()}, obj); err != nil {
		t.Errorf("a text that is not written was reported: %v", err)
	}

	asJSON := dataFileWritten("data", true)
	jobj := map[string]any{"data_file": map[string]any{"id": 3.0}, "data": map[string]any{"a": 1.0, "b": []any{"x"}}}
	same := fwresource.Values{"data_file_id": types.Int64Value(3), "data": types.StringValue(`{ "b": ["x"], "a": 1 }`)}
	if err := asJSON(same, jobj); err != nil {
		t.Errorf("the same JSON in another layout was reported: %v", err)
	}
	other := fwresource.Values{"data_file_id": types.Int64Value(3), "data": types.StringValue(`{"a": 2}`)}
	if err := asJSON(other, jobj); err == nil {
		t.Error("another JSON was not reported")
	}
}

// NetBox 4.7.5 replaces the text of a template with the content of its data file. The error says so,
// for a create and for an update.
func TestEngineReportsContentThatNetBoxReplaced(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_config_template"
	fake.rewrite = map[string]func(map[string]any){
		"extras/config-templates": func(obj map[string]any) {
			if obj["data_file"] != nil {
				obj["template_code"] = "FILE\n"
				obj["data_path"] = "t.j2"
			}
		},
	}
	set := map[string]tftypes.Value{
		"name":          str("t"),
		"template_code": str("MINE\n"),
		"data_file_id":  tftypes.NewValue(tftypes.Number, 3),
	}
	diags, _ := h.tryApply(name, nullObject(), set)
	h.wantError("a create that NetBox changed", diags)
	found := false
	for _, d := range diags {
		if strings.Contains(d.Summary, "NetBox changed what was written") && strings.Contains(d.Detail, "t.j2") {
			found = true
		}
	}
	if !found {
		t.Errorf("the error does not explain what changed: %v", diags)
	}

	// an object that exists without a file, then linked to one by an update
	delete(fake.rewrite, "extras/config-templates")
	plain := h.apply(name, nullObject(), map[string]tftypes.Value{"name": str("u"), "template_code": str("MINE\n")})
	fake.rewrite["extras/config-templates"] = func(obj map[string]any) {
		if obj["data_file"] != nil {
			obj["template_code"] = "FILE\n"
			obj["data_path"] = "t.j2"
		}
	}
	diags, _ = h.tryApply(name, plain, map[string]tftypes.Value{
		"name":          str("u"),
		"template_code": str("MINE\n"),
		"data_file_id":  tftypes.NewValue(tftypes.Number, 3),
	})
	h.wantError("an update that NetBox changed", diags)
}

func TestDataFileLinkPlanNeedsADataSourceAttribute(t *testing.T) {
	plan := fwresource.Values{
		"last_updated": types.StringUnknown(),
		"data_file_id": types.Int64Value(3),
	}
	if got := dataFileLinkPlan(plan, fwresource.Values{}); got != nil {
		t.Errorf("a plan without data_source_id was changed: %v", got)
	}
}

func TestChoiceSetOrderIgnoresListsItCannotRead(t *testing.T) {
	str := func(s string) attr.Value { return types.StringValue(s) }
	inner := func(vals ...attr.Value) attr.Value { return types.ListValueMust(types.StringType, vals) }
	list := func(items ...attr.Value) attr.Value {
		return types.ListValueMust(types.ListType{ElemType: types.StringType}, items)
	}
	on := types.BoolValue(true)
	for name, choices := range map[string]attr.Value{
		"an empty choice":     list(inner()),
		"an unknown choice":   list(types.ListUnknown(types.StringType)),
		"a null value inside": list(inner(types.StringNull(), str("L"))),
		"an unknown value":    list(inner(types.StringUnknown(), str("L"))),
	} {
		cfg := fwresource.Values{"order_alphabetically": on, "extra_choices": choices}
		if diags := choiceSetOrder(context.Background(), cfg); diags.HasError() {
			t.Errorf("%s was reported: %v", name, diags)
		}
	}
}

// A PUT without master makes NetBox clear the master of a virtual chassis, so updating the chassis
// has to be a PATCH. The master is set from the device or from netbox_virtual_chassis_master.
func TestVirtualChassisUpdateKeepsTheMaster(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	const name = "netbox_virtual_chassis"
	created := h.apply(name, nullObject(), map[string]tftypes.Value{"name": str("vc")})
	h.apply(name, created, map[string]tftypes.Value{"name": str("vc2")})
	for _, call := range fake.calls {
		if strings.HasPrefix(call, "PUT ") && strings.Contains(call, "virtual-chassis") {
			t.Errorf("the chassis was updated with %s, which clears its master", call)
		}
	}
}

// Several providers can be created at the same time, for example by the acceptance tests.
func TestDataSourcesAreSafeToListConcurrently(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			(&Provider{}).DataSources(context.Background())
		}()
	}
	wg.Wait()
}
