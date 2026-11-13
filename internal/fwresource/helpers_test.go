package fwresource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNumberHelpers(t *testing.T) {
	for _, tt := range []struct {
		in   any
		want int64
	}{{3, 3}, {int64(4), 4}, {5.9, 5}, {"6", 6}, {"x", 0}, {nil, 0}} {
		if got := toInt64(tt.in); got != tt.want {
			t.Errorf("toInt64(%v) = %d, want %d", tt.in, got, tt.want)
		}
	}
	for _, tt := range []struct {
		in   any
		want float64
	}{{1.5, 1.5}, {2, 2}, {int64(3), 3}, {"4.5", 4.5}, {"x", 0}, {nil, 0}} {
		if got := toFloat64(tt.in); got != tt.want {
			t.Errorf("toFloat64(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestNullOf(t *testing.T) {
	for kind, want := range map[Kind]attr.Value{
		Int:       types.Int64Null(),
		Float:     types.Float64Null(),
		Bool:      types.BoolNull(),
		String:    types.StringNull(),
		StringMap: types.MapNull(types.StringType),
		StringSet: types.SetNull(types.StringType),
		IntList:   types.ListNull(types.Int64Type),
	} {
		if got := nullOf(Attr{Kind: kind}); !got.Equal(want) {
			t.Errorf("kind %d: %s, want %s", kind, got, want)
		}
	}
}

func TestQueryValue(t *testing.T) {
	set, _ := types.SetValue(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	list, _ := types.ListValue(types.Int64Type, []attr.Value{types.Int64Value(1), types.Int64Value(2)})
	for _, tt := range []struct {
		in   attr.Value
		want int
	}{
		{types.StringValue("x"), 1}, {types.StringValue(""), 0}, {types.StringNull(), 0},
		{types.Int64Value(3), 1}, {types.Int64Value(0), 0}, {types.Float64Value(1.5), 1},
		{types.BoolValue(false), 1}, {types.BoolNull(), 0}, {set, 2}, {list, 2}, {nil, 0},
	} {
		if got := queryValue(tt.in); len(got) != tt.want {
			t.Errorf("queryValue(%v) = %v, want %d values", tt.in, got, tt.want)
		}
	}
}

func TestRequiredWithDescription(t *testing.T) {
	v := requiredWith{attr: "a", other: "b"}
	if v.Description(context.Background()) != "a requires b" || v.MarkdownDescription(context.Background()) != "a requires b" {
		t.Error("description")
	}
}
