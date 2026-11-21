package fwresource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func customFieldsValue(m map[string]string) types.Map {
	elems := map[string]attr.Value{}
	for k, v := range m {
		elems[k] = types.StringValue(v)
	}
	return types.MapValueMust(types.StringType, elems)
}

// A custom field that is in the state but not in the configuration, as after an upgrade from the
// SDK provider, must not be sent, so NetBox keeps its value.
func TestCustomFieldMissingFromConfigIsNotCleared(t *testing.T) {
	r := &Resource{spec: (&Spec{TypeName: "netbox_x", CustomFields: true}).withZeroDefaults()}
	plan := Values{customFieldsKey: customFieldsValue(map[string]string{"kept": "1"})}
	prior := Values{customFieldsKey: customFieldsValue(map[string]string{"kept": "1", "environment": "prod"})}
	body, err := r.body(context.Background(), plan, prior)
	if err != nil {
		t.Fatal(err)
	}
	cf, _ := body["custom_fields"].(map[string]any)
	if _, sent := cf["environment"]; sent {
		t.Errorf("a field that is not in the configuration was sent: %v", cf)
	}
	if cf["kept"] != "1" {
		t.Errorf("a configured field was not sent: %v", cf)
	}

	// an empty string still clears a field
	plan = Values{customFieldsKey: customFieldsValue(map[string]string{"environment": ""})}
	body, _ = r.body(context.Background(), plan, prior)
	cf, _ = body["custom_fields"].(map[string]any)
	if v, sent := cf["environment"]; !sent || v != "" {
		t.Errorf("an empty string was not sent: %v", cf)
	}
}

func TestGroupValidatorsIgnoreUnsetZeroValues(t *testing.T) {
	obj := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.Number, "b": tftypes.Number, "c": tftypes.Bool}}
	check := func(v resource.ConfigValidator, a, b, c tftypes.Value) bool {
		cfg := tfsdk.Config{Raw: tftypes.NewValue(obj, map[string]tftypes.Value{"a": a, "b": b, "c": c})}
		var resp resource.ValidateConfigResponse
		v.ValidateResource(context.Background(), resource.ValidateConfigRequest{Config: cfg}, &resp)
		return resp.Diagnostics.HasError()
	}
	num := func(n int64) tftypes.Value { return tftypes.NewValue(tftypes.Number, n) }
	null := tftypes.NewValue(tftypes.Number, nil)
	nullBool := tftypes.NewValue(tftypes.Bool, nil)
	zero := map[string]bool{"a": true, "b": true, "c": true}

	conflict := groupValidator{kind: groupConflicts, names: []string{"a", "b"}, zero: zero}
	if check(conflict, num(5), num(0), nullBool) {
		t.Error("0 counted as set in a conflict")
	}
	if !check(conflict, num(5), num(7), nullBool) {
		t.Error("two values did not conflict")
	}
	if !check(groupValidator{kind: groupConflicts, names: []string{"a", "b"}}, num(5), num(0), nullBool) {
		t.Error("0 of an attribute without a zero default must count as set")
	}

	exactly := groupValidator{kind: groupExactlyOne, names: []string{"a", "b"}, zero: zero}
	if check(exactly, num(0), num(3), nullBool) || !check(exactly, num(0), num(0), nullBool) || !check(exactly, num(1), num(3), nullBool) {
		t.Error("exactly one")
	}
	atLeast := groupValidator{kind: groupAtLeastOne, names: []string{"a", "b"}, zero: zero}
	if check(atLeast, null, num(3), nullBool) || !check(atLeast, num(0), null, nullBool) {
		t.Error("at least one")
	}

	// false and 0 do not make the other attribute required
	req := requiredWith{attr: "c", other: "a", zero: zero}
	if check(req, num(0), null, tftypes.NewValue(tftypes.Bool, false)) {
		t.Error("false required another attribute")
	}
	if !check(req, null, null, tftypes.NewValue(tftypes.Bool, true)) {
		t.Error("true did not require another attribute")
	}
	if check(req, num(4), null, tftypes.NewValue(tftypes.Bool, true)) {
		t.Error("required attribute was reported missing")
	}
}

func TestZeroIsNotRangeChecked(t *testing.T) {
	a := Attr{Name: "n", Kind: Int, Optional: true, Min: i64(1), Max: i64(10), zero: true}
	vs := a.schemaAttribute().(schema.Int64Attribute).Validators
	run := func(n int64) bool {
		var resp validator.Int64Response
		for _, v := range vs {
			v.ValidateInt64(context.Background(), validator.Int64Request{Path: path.Root("n"), ConfigValue: types.Int64Value(n)}, &resp)
		}
		return resp.Diagnostics.HasError()
	}
	if run(0) || run(5) || !run(11) {
		t.Error("range check")
	}
	// an attribute without a zero default is checked as usual
	a.zero = false
	vs = a.schemaAttribute().(schema.Int64Attribute).Validators
	if !run(0) {
		t.Error("0 was accepted for an attribute that is not zero by default")
	}
}

func i64(n int64) *int64 { return &n }

// A write-only flag that is not set is left out of the request, because NetBox rejects null for it.
func TestBodyLeavesOutAWriteOnlyValueThatIsNotSet(t *testing.T) {
	spec := (&Spec{TypeName: "netbox_x", Attrs: []Attr{
		{Name: "flag", Kind: Bool, Optional: true, WriteOnly: true},
	}}).withZeroDefaults()
	r := &Resource{spec: spec}
	prior := Values{"flag": types.BoolNull()}
	body, err := r.body(context.Background(), Values{"flag": types.BoolNull()}, prior)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := body["flag"]; ok {
		t.Errorf("a null write-only flag was sent: %v", body)
	}
	body, err = r.body(context.Background(), Values{"flag": types.BoolValue(true)}, prior)
	if err != nil || body["flag"] != true {
		t.Errorf("a set write-only flag was not sent: %v %v", body, err)
	}
}
