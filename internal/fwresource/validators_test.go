package fwresource

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestGroupValidatorDescriptions(t *testing.T) {
	ctx := context.Background()
	for kind, want := range map[groupKind]string{
		groupConflicts:  "at most one of a, b",
		groupExactlyOne: "exactly one of a, b",
		groupAtLeastOne: "at least one of a, b",
	} {
		v := groupValidator{kind: kind, names: []string{"a", "b"}}
		if got := v.Description(ctx); !strings.Contains(got, want) || v.MarkdownDescription(ctx) != got {
			t.Errorf("kind %d: got %q, want %q", kind, got, want)
		}
	}
}

func TestGroupValidatorEdgeCases(t *testing.T) {
	obj := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String, "b": tftypes.String}}
	run := func(v groupValidator, raw tftypes.Value) bool {
		var resp resource.ValidateConfigResponse
		v.ValidateResource(context.Background(), resource.ValidateConfigRequest{Config: tfsdk.Config{Raw: raw}}, &resp)
		return resp.Diagnostics.HasError()
	}
	val := func(a, b tftypes.Value) tftypes.Value {
		return tftypes.NewValue(obj, map[string]tftypes.Value{"a": a, "b": b})
	}
	null := tftypes.NewValue(tftypes.String, nil)
	unknown := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	text := tftypes.NewValue(tftypes.String, "x")

	for _, kind := range []groupKind{groupConflicts, groupExactlyOne, groupAtLeastOne} {
		v := groupValidator{kind: kind, names: []string{"a", "b"}}
		// an unknown value means the group cannot be judged yet
		if run(v, val(unknown, null)) {
			t.Errorf("kind %d: an unknown value was judged", kind)
		}
		// a configuration that is not an object is left to Terraform
		if run(v, tftypes.NewValue(tftypes.String, "x")) {
			t.Errorf("kind %d: a value that is not an object was judged", kind)
		}
	}
	// names that are not in the configuration are skipped
	if !run(groupValidator{kind: groupAtLeastOne, names: []string{"missing", "a"}}, val(null, null)) {
		t.Error("nothing set did not fail")
	}
	if run(groupValidator{kind: groupAtLeastOne, names: []string{"missing", "a"}}, val(text, null)) {
		t.Error("a set attribute failed")
	}
	if !run(groupValidator{kind: groupExactlyOne, names: []string{"a", "b"}}, val(null, null)) ||
		!run(groupValidator{kind: groupExactlyOne, names: []string{"a", "b"}}, val(text, text)) {
		t.Error("exactly one")
	}
}

func TestIsSetValue(t *testing.T) {
	str := func(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }
	setT := tftypes.Set{ElementType: tftypes.String}
	listT := tftypes.List{ElementType: tftypes.String}
	mapT := tftypes.Map{ElementType: tftypes.String}
	for name, tt := range map[string]struct {
		v    tftypes.Value
		zero bool
		want bool
	}{
		"empty string, zero":       {str(""), true, false},
		"empty string, not zero":   {str(""), false, true},
		"text":                     {str("x"), true, true},
		"null":                     {tftypes.NewValue(tftypes.String, nil), false, false},
		"number zero":              {tftypes.NewValue(tftypes.Number, 0), true, false},
		"number":                   {tftypes.NewValue(tftypes.Number, 3), true, true},
		"false":                    {tftypes.NewValue(tftypes.Bool, false), true, false},
		"true":                     {tftypes.NewValue(tftypes.Bool, true), true, true},
		"empty set":                {tftypes.NewValue(setT, []tftypes.Value{}), true, false},
		"set":                      {tftypes.NewValue(setT, []tftypes.Value{str("a")}), true, true},
		"empty list":               {tftypes.NewValue(listT, []tftypes.Value{}), true, false},
		"list":                     {tftypes.NewValue(listT, []tftypes.Value{str("a")}), true, true},
		"empty map":                {tftypes.NewValue(mapT, map[string]tftypes.Value{}), true, false},
		"map":                      {tftypes.NewValue(mapT, map[string]tftypes.Value{"k": str("v")}), true, true},
		"other type with zero set": {tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}), true, true},
	} {
		if got := isSetValue(tt.v, tt.zero); got != tt.want {
			t.Errorf("%s: got %v, want %v", name, got, tt.want)
		}
	}
}

// A switch that is off does not need the attribute it depends on.
func TestRequiredWithIgnoresASwitchThatIsOff(t *testing.T) {
	obj := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"on": tftypes.Bool, "file": tftypes.Number}}
	run := func(on tftypes.Value) bool {
		cfg := tfsdk.Config{Raw: tftypes.NewValue(obj, map[string]tftypes.Value{"on": on, "file": tftypes.NewValue(tftypes.Number, nil)})}
		var resp resource.ValidateConfigResponse
		requiredWith{attr: "on", other: "file"}.ValidateResource(context.Background(), resource.ValidateConfigRequest{Config: cfg}, &resp)
		return resp.Diagnostics.HasError()
	}
	if run(tftypes.NewValue(tftypes.Bool, false)) {
		t.Error("false required another attribute")
	}
	if !run(tftypes.NewValue(tftypes.Bool, true)) {
		t.Error("true did not require another attribute")
	}
}

func TestSameValuesRejectsDifferentValues(t *testing.T) {
	setT := tftypes.Set{ElementType: tftypes.String}
	obj := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": setT, "b": setT}}
	val := func(a, b []string) tftypes.Value {
		mk := func(v []string) tftypes.Value {
			if v == nil {
				return tftypes.NewValue(setT, nil)
			}
			elems := make([]tftypes.Value, len(v))
			for i, s := range v {
				elems[i] = tftypes.NewValue(tftypes.String, s)
			}
			return tftypes.NewValue(setT, elems)
		}
		return tftypes.NewValue(obj, map[string]tftypes.Value{"a": mk(a), "b": mk(b)})
	}
	run := func(a, b []string) bool {
		var resp resource.ValidateConfigResponse
		sameValues{a: "a", b: "b"}.ValidateResource(context.Background(), resource.ValidateConfigRequest{Config: tfsdk.Config{Raw: val(a, b)}}, &resp)
		return resp.Diagnostics.HasError()
	}
	if run([]string{"x", "y"}, []string{"y", "x"}) {
		t.Error("the same values in another order were rejected")
	}
	if !run([]string{"x"}, []string{"y"}) {
		t.Error("different values were accepted")
	}
	if run([]string{"x"}, nil) || run(nil, nil) {
		t.Error("an attribute that is not written was compared")
	}
}
