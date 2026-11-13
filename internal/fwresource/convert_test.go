package fwresource

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestScalarFromJSON(t *testing.T) {
	if s, ok := stringFromJSON(nil); ok || s != "" {
		t.Error("nil is not a string")
	}
	for _, tt := range []struct {
		in   any
		want string
	}{
		{"x", "x"},
		{true, "true"},
		{float64(3), "3"},
		{1.5, "1.5"},
		{map[string]any{"value": "active", "label": "Active"}, "active"},
		{map[string]any{"id": float64(7)}, "7"},
		{map[string]any{"a": 1}, `{"a":1}`},
		{[]any{"a", "b"}, `["a","b"]`},
	} {
		if got, ok := stringFromJSON(tt.in); !ok || got != tt.want {
			t.Errorf("stringFromJSON(%v) = %q, %v; want %q", tt.in, got, ok, tt.want)
		}
	}

	for _, tt := range []struct {
		in   any
		want int64
		ok   bool
	}{
		{float64(4), 4, true},
		{float64(4.6), 5, true},
		{"12", 12, true},
		{"x", 0, false},
		{map[string]any{"id": float64(9)}, 9, true},
		{map[string]any{"value": float64(2)}, 2, true},
		{true, 0, false},
		{nil, 0, false},
	} {
		if got, ok := intFromJSON(tt.in); got != tt.want || ok != tt.ok {
			t.Errorf("intFromJSON(%v) = %d, %v", tt.in, got, ok)
		}
	}

	if f, ok := floatFromJSON("2.5"); !ok || f != 2.5 {
		t.Error("float from string")
	}
	if f, ok := floatFromJSON(float64(1)); !ok || f != 1 {
		t.Error("float from float")
	}
	if _, ok := floatFromJSON("x"); ok {
		t.Error("float from garbage")
	}
	if _, ok := floatFromJSON(nil); ok {
		t.Error("float from nil")
	}
	if b, ok := boolFromJSON(true); !ok || !b {
		t.Error("bool")
	}
	if b, ok := boolFromJSON(map[string]any{"value": true}); !ok || !b {
		t.Error("bool from choice")
	}
	if _, ok := boolFromJSON("yes"); ok {
		t.Error("bool from string")
	}
}

func TestFromJSONKinds(t *testing.T) {
	str := func(s string) attr.Value { return types.StringValue(s) }
	for _, tt := range []struct {
		name  string
		a     Attr
		in    any
		prior attr.Value
		want  attr.Value
	}{
		{"string", Attr{Kind: String, Optional: true}, "x", nil, str("x")},
		{"unset optional string", Attr{Kind: String, Optional: true}, "", nil, types.StringNull()},
		{"cleared string", Attr{Kind: String, Optional: true}, "", str("old"), str("")},
		{"required empty", Attr{Kind: String, Required: true}, "", nil, str("")},
		{"trimmed", Attr{Kind: String, Optional: true}, "x", str(" x "), str(" x ")},
		{"ip only", Attr{Kind: String, IPOnly: true}, "10.0.0.1/24", nil, str("10.0.0.1")},
		{"ip only invalid", Attr{Kind: String, IPOnly: true}, "nonsense", nil, str("")},
		{"case insensitive", Attr{Kind: String, Optional: true, CaseInsensitive: true}, "ABC", str("abc"), str("abc")},
		{"time same instant", Attr{Kind: String, Optional: true, Time: true}, "2030-01-01T01:00:00+01:00", str("2030-01-01T00:00:00Z"), str("2030-01-01T00:00:00Z")},
		{"time normalised", Attr{Kind: String, Optional: true, Time: true}, "2030-01-01T00:00:00Z", nil, str("2030-01-01T00:00:00.000Z")},
		{"loose json nil", Attr{Kind: String, LooseJSON: true}, nil, nil, str("")},
		{"loose json object", Attr{Kind: String, LooseJSON: true}, map[string]any{"a": 1.0}, nil, str(`{"a":1}`)},
		{"loose json keeps prior", Attr{Kind: String, LooseJSON: true}, map[string]any{"a": 1.0}, str(`{ "a": 1 }`), str(`{ "a": 1 }`)},
		{"loose json string", Attr{Kind: String, LooseJSON: true}, "plain", nil, str("plain")},
		{"json nil", Attr{Kind: JSON}, nil, nil, types.StringNull()},
		{"json empty object", Attr{Kind: JSON}, map[string]any{}, nil, types.StringNull()},
		{"json empty list", Attr{Kind: JSON}, []any{}, nil, types.StringNull()},
		{"json", Attr{Kind: JSON}, map[string]any{"a": 1.0}, nil, str(`{"a":1}`)},
		{"json keeps prior", Attr{Kind: JSON}, map[string]any{"a": 1.0}, str(`{"a": 1}`), str(`{"a": 1}`)},
		{"int", Attr{Kind: Int}, float64(3), nil, types.Int64Value(3)},
		{"int missing", Attr{Kind: Int}, nil, nil, types.Int64Null()},
		{"int default", Attr{Kind: Int, Default: int64(5)}, nil, nil, types.Int64Value(5)},
		{"int zero", Attr{Kind: Int, Zero: true}, nil, nil, types.Int64Value(0)},
		{"float", Attr{Kind: Float}, 2.5, nil, types.Float64Value(2.5)},
		{"float missing", Attr{Kind: Float}, nil, nil, types.Float64Null()},
		{"float zero", Attr{Kind: Float, Zero: true}, nil, nil, types.Float64Value(0)},
		{"bool", Attr{Kind: Bool, Optional: true}, true, nil, types.BoolValue(true)},
		{"bool false unset", Attr{Kind: Bool, Optional: true}, false, nil, types.BoolNull()},
		{"bool missing", Attr{Kind: Bool}, nil, nil, types.BoolNull()},
		{"bool default", Attr{Kind: Bool, Default: true}, nil, nil, types.BoolValue(true)},
		{"bool zero", Attr{Kind: Bool, Zero: true}, nil, nil, types.BoolValue(false)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.fromJSON(tt.in, tt.prior)
			if !got.Equal(tt.want) {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFromJSONCollections(t *testing.T) {
	strs := func(s ...string) []attr.Value {
		out := make([]attr.Value, len(s))
		for i, v := range s {
			out[i] = types.StringValue(v)
		}
		return out
	}

	set := Attr{Kind: StringSet, Optional: true}.fromJSON([]any{"b", "a"}, nil)
	want, _ := types.SetValue(types.StringType, strs("a", "b"))
	if !set.Equal(want) {
		t.Errorf("set = %s", set)
	}
	if !(Attr{Kind: StringSet, Optional: true}).fromJSON([]any{}, nil).IsNull() {
		t.Error("an empty unset set should be null")
	}
	if (Attr{Kind: StringList, Optional: true}).fromJSON([]any{}, types.ListNull(types.StringType)).IsNull() == false {
		t.Error("an empty unset list should be null")
	}

	tags := Attr{Kind: StringSet, ElemKey: "name"}.fromJSON([]any{map[string]any{"name": "t1"}, "t2"}, nil)
	if len(tags.(types.Set).Elements()) != 2 {
		t.Errorf("tags = %s", tags)
	}

	ints := Attr{Kind: IntList, Ref: true}.fromJSON([]any{float64(1), map[string]any{"id": float64(2)}}, nil)
	if len(ints.(types.List).Elements()) != 2 {
		t.Errorf("ints = %s", ints)
	}
	intSet := Attr{Kind: IntSet}.fromJSON([]any{float64(1)}, nil)
	if len(intSet.(types.Set).Elements()) != 1 {
		t.Errorf("int set = %s", intSet)
	}

	lists := Attr{Kind: IntLists}.fromJSON([]any{[]any{float64(1), float64(2)}, []any{float64(3)}}, nil)
	if len(lists.(types.List).Elements()) != 2 {
		t.Errorf("int lists = %s", lists)
	}
	slists := Attr{Kind: StringLists}.fromJSON([]any{[]any{"a"}}, nil)
	if len(slists.(types.List).Elements()) != 1 {
		t.Errorf("string lists = %s", slists)
	}
	if !(Attr{Kind: IntLists, Optional: true}).fromJSON(nil, nil).IsNull() {
		t.Error("empty lists of lists should be null")
	}

	m := Attr{Kind: StringMap}.fromJSON(map[string]any{"a": "1", "b": float64(2), "c": nil}, nil)
	if len(m.(types.Map).Elements()) != 2 {
		t.Errorf("map = %s", m)
	}

	nested := Attr{Kind: NestedList, Sub: []Attr{{Name: "start", Kind: Int}, {Name: "end", Kind: Int}}, Pairs: true}
	got := nested.fromJSON([]any{[]any{float64(1), float64(5)}, []any{float64(9)}}, nil)
	if n := len(got.(types.List).Elements()); n != 1 {
		t.Errorf("pairs = %s", got)
	}
	single := Attr{Kind: NestedSet, Computed: true, Sub: []Attr{{Name: "id", Kind: Int}, {Name: "name", Kind: String}}}.
		fromJSON(map[string]any{"id": float64(1), "name": "n"}, nil)
	if n := len(single.(types.Set).Elements()); n != 1 {
		t.Errorf("single nested object = %s", single)
	}
	if !(Attr{Kind: NestedList, Sub: []Attr{{Name: "id", Kind: Int}}}).fromJSON(nil, nil).IsNull() {
		t.Error("no nested objects should be null")
	}
}

func TestToJSON(t *testing.T) {
	zeroed := func(a Attr) Attr { a.zero = true; return a }
	prior := types.StringValue("old")
	for _, tt := range []struct {
		name   string
		a      Attr
		val    attr.Value
		prior  attr.Value
		want   any
		wantOK bool
	}{
		{"unknown", Attr{Kind: String}, types.StringUnknown(), nil, nil, false},
		{"string", Attr{Kind: String}, types.StringValue("x"), nil, "x", true},
		{"null on create", Attr{Kind: String}, types.StringNull(), nil, nil, false},
		{"null on update", Attr{Kind: String}, types.StringNull(), prior, "", true},
		{"null nullable on update", Attr{Kind: String, Nullable: true}, types.StringNull(), prior, nil, true},
		{"null set on update", Attr{Kind: StringSet}, types.SetNull(types.StringType), prior, []any{}, true},
		{"null int on update", Attr{Kind: Int}, types.Int64Null(), prior, nil, true},
		{"computed null", Attr{Kind: String, Computed: true}, types.StringNull(), prior, nil, false},
		{"int", Attr{Kind: Int}, types.Int64Value(3), nil, int64(3), true},
		{"float", Attr{Kind: Float}, types.Float64Value(1.5), nil, 1.5, true},
		{"bool", Attr{Kind: Bool}, types.BoolValue(true), nil, true, true},
		{"nullable zero", zeroed(Attr{Kind: Int, Nullable: true}), types.Int64Value(0), nil, nil, true},
		{"zero ref on create", Attr{Kind: Int, Ref: true}, types.Int64Value(0), nil, nil, false},
		{"zero ref on update", Attr{Kind: Int, Ref: true}, types.Int64Value(0), prior, nil, true},
		{"zero string on create", zeroed(Attr{Kind: String}), types.StringValue(""), nil, nil, false},
		{"zero string cleared", zeroed(Attr{Kind: String}), types.StringValue(""), types.StringValue("was"), "", true},
		{"zero string still zero", zeroed(Attr{Kind: String}), types.StringValue(""), types.StringValue(""), nil, false},
		{"zero int sent", zeroed(Attr{Kind: Int}), types.Int64Value(0), nil, int64(0), true},
		{"zero bool sent", zeroed(Attr{Kind: Bool}), types.BoolValue(false), nil, false, true},
		{"json", Attr{Kind: JSON}, types.StringValue(`{"a":1}`), nil, map[string]any{"a": float64(1)}, true},
		{"loose json", Attr{Kind: String, LooseJSON: true}, types.StringValue(`[1]`), nil, []any{float64(1)}, true},
		{"loose json plain", Attr{Kind: String, LooseJSON: true}, types.StringValue(`plain`), nil, "plain", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := tt.a.toJSON(tt.val, tt.prior)
			if err != nil {
				t.Fatal(err)
			}
			if ok != tt.wantOK {
				t.Fatalf("sent = %v, want %v", ok, tt.wantOK)
			}
			if ok && !jsonEqual(mustJSON(got), mustJSON(tt.want)) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

	if _, _, err := (Attr{Name: "x", Kind: JSON}).toJSON(types.StringValue("{"), nil); err == nil {
		t.Error("invalid JSON was accepted")
	}

	set, _ := types.SetValue(types.StringType, []attr.Value{types.StringValue("a")})
	if got, _, _ := (Attr{Kind: StringSet}).toJSON(set, nil); len(got.([]any)) != 1 {
		t.Errorf("set = %v", got)
	}
	ints, _ := types.ListValue(types.Int64Type, []attr.Value{types.Int64Value(1)})
	if got, _, _ := (Attr{Kind: IntList}).toJSON(ints, nil); len(got.([]any)) != 1 {
		t.Errorf("ints = %v", got)
	}
	inner, _ := types.ListValue(types.Int64Type, []attr.Value{types.Int64Value(1), types.Int64Value(2)})
	lists, _ := types.ListValue(types.ListType{ElemType: types.Int64Type}, []attr.Value{inner})
	if got, _, _ := (Attr{Kind: IntLists}).toJSON(lists, nil); len(got.([]any)) != 1 {
		t.Errorf("lists = %v", got)
	}
	sinner, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("a")})
	slists, _ := types.ListValue(types.ListType{ElemType: types.StringType}, []attr.Value{sinner})
	if got, _, _ := (Attr{Kind: StringLists}).toJSON(slists, nil); len(got.([]any)) != 1 {
		t.Errorf("string lists = %v", got)
	}

	nested := Attr{Kind: NestedList, Sub: []Attr{{Name: "start", Kind: Int}, {Name: "end", Kind: Int}}}
	obj, _ := types.ObjectValue(nested.objectType().AttrTypes, map[string]attr.Value{"start": types.Int64Value(1), "end": types.Int64Value(2)})
	nl, _ := types.ListValue(nested.objectType(), []attr.Value{obj})
	got, _, _ := nested.toJSON(nl, nil)
	if m := got.([]any)[0].(map[string]any); m["start"] != int64(1) || m["end"] != int64(2) {
		t.Errorf("nested = %v", got)
	}
}

func TestIsZeroAndJSONEqual(t *testing.T) {
	empty, _ := types.SetValue(types.StringType, nil)
	emptyList, _ := types.ListValue(types.StringType, nil)
	for _, v := range []attr.Value{nil, types.StringNull(), types.StringValue(""), types.Int64Value(0), types.Float64Value(0), types.BoolValue(false), empty, emptyList} {
		if !isZero(v) {
			t.Errorf("%v should be zero", v)
		}
	}
	for _, v := range []attr.Value{types.StringValue("x"), types.Int64Value(1), types.Float64Value(1), types.BoolValue(true)} {
		if isZero(v) {
			t.Errorf("%v should not be zero", v)
		}
	}
	if isZero(types.MapValueMust(types.StringType, map[string]attr.Value{"a": types.StringValue("b")})) {
		t.Error("a map is never zero")
	}

	if !jsonEqual(`{"a":1,"b":2}`, `{ "b": 2, "a": 1 }`) {
		t.Error("same document, different spelling")
	}
	if jsonEqual(`{"a":1}`, `{"a":2}`) {
		t.Error("different documents")
	}
	if !jsonEqual(" x ", "x") || jsonEqual("x", "y") {
		t.Error("non-JSON compares trimmed")
	}
}

func TestPickAndValueType(t *testing.T) {
	obj := map[string]any{
		"name":  "n",
		"owner": map[string]any{"id": float64(4)},
		"list":  []any{map[string]any{"v": "first"}},
	}
	if got := (Attr{Name: "name"}).pick(obj); got != "n" {
		t.Errorf("pick = %v", got)
	}
	if got := (Attr{Name: "x", Path: []string{"owner", "id"}}).pick(obj); got != float64(4) {
		t.Errorf("nested pick = %v", got)
	}
	if got := (Attr{Name: "x", Path: []string{"list", "[0]", "v"}}).pick(obj); got != "first" {
		t.Errorf("list pick = %v", got)
	}
	if got := (Attr{Name: "x", Path: []string{"missing", "id"}}).pick(obj); got != nil {
		t.Errorf("missing pick = %v", got)
	}

	for kind, want := range map[Kind]attr.Type{
		Int:         types.Int64Type,
		Bool:        types.BoolType,
		Float:       types.Float64Type,
		String:      types.StringType,
		StringMap:   types.MapType{ElemType: types.StringType},
		StringSet:   types.SetType{ElemType: types.StringType},
		IntSet:      types.SetType{ElemType: types.Int64Type},
		StringList:  types.ListType{ElemType: types.StringType},
		IntList:     types.ListType{ElemType: types.Int64Type},
		IntLists:    types.ListType{ElemType: types.ListType{ElemType: types.Int64Type}},
		StringLists: types.ListType{ElemType: types.ListType{ElemType: types.StringType}},
	} {
		if got := (Attr{Kind: kind}).valueType(); !got.Equal(want) {
			t.Errorf("kind %d: type %s, want %s", kind, got, want)
		}
	}
	nested := Attr{Kind: NestedSet, Sub: []Attr{{Name: "a", Kind: Int}}}
	if !nested.valueType().Equal(types.SetType{ElemType: nested.objectType()}) {
		t.Error("nested set type")
	}
	nested.Kind = NestedList
	if !nested.valueType().Equal(types.ListType{ElemType: nested.objectType()}) {
		t.Error("nested list type")
	}
	if !(Attr{Kind: StringSet}).nullCollection().IsNull() || !(Attr{Kind: IntList}).nullCollection().IsNull() {
		t.Error("null collections")
	}
}

func mustJSON(v any) string {
	if s, ok := v.(string); ok {
		return `"` + s + `"`
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func TestConvertEdgeCases(t *testing.T) {
	if s, ok := stringFromJSON(make(chan int)); ok || s != "" {
		t.Errorf("an unencodable value gave %q, %v", s, ok)
	}
	if got := (Attr{Kind: Kind(99)}).fromJSON("x", nil); got != nil {
		t.Errorf("an unknown kind gave %v", got)
	}
	if _, _, err := (Attr{Name: "m", Kind: String}).toJSON(types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}), nil); err == nil {
		t.Error("a value of an unsupported type was accepted")
	}
}

func TestStringMapToJSON(t *testing.T) {
	a := Attr{Name: "m", Kind: StringMap, Optional: true}
	got, ok, err := a.toJSON(types.MapValueMust(types.StringType, map[string]attr.Value{"a": types.StringValue("b")}), nil)
	if err != nil || !ok {
		t.Fatalf("a map was not sent: %v %v", ok, err)
	}
	if m, isMap := got.(map[string]any); !isMap || m["a"] != "b" {
		t.Errorf("got %v", got)
	}

	// a map that was removed from the configuration is cleared on update
	cleared, ok, err := a.toJSON(types.MapNull(types.StringType), types.MapValueMust(types.StringType, map[string]attr.Value{"a": types.StringValue("b")}))
	if err != nil || !ok {
		t.Fatalf("a removed map was not cleared: %v %v", ok, err)
	}
	if m, isMap := cleared.(map[string]any); !isMap || len(m) != 0 {
		t.Errorf("got %v", cleared)
	}
}
