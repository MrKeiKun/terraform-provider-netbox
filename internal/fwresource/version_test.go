package fwresource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIsSet(t *testing.T) {
	str := types.StringType
	set := func(vals ...string) attr.Value {
		var elems []attr.Value
		for _, v := range vals {
			elems = append(elems, types.StringValue(v))
		}
		out, _ := types.SetValue(str, elems)
		return out
	}
	list := func(vals ...string) attr.Value {
		var elems []attr.Value
		for _, v := range vals {
			elems = append(elems, types.StringValue(v))
		}
		out, _ := types.ListValue(str, elems)
		return out
	}
	mapOf := func(vals map[string]string) attr.Value {
		elems := map[string]attr.Value{}
		for k, v := range vals {
			elems[k] = types.StringValue(v)
		}
		out, _ := types.MapValue(str, elems)
		return out
	}

	for name, tt := range map[string]struct {
		value attr.Value
		want  bool
	}{
		"nothing":       {nil, false},
		"null":          {types.StringNull(), false},
		"unknown":       {types.StringUnknown(), false},
		"a string":      {types.StringValue(""), true},
		"an empty set":  {set(), false},
		"a set":         {set("a"), true},
		"an empty list": {list(), false},
		"a list":        {list("a"), true},
		"an empty map":  {mapOf(nil), false},
		"a map":         {mapOf(map[string]string{"a": "b"}), true},
	} {
		if got := isSet(tt.value); got != tt.want {
			t.Errorf("%s: isSet = %v, want %v", name, got, tt.want)
		}
	}
}

func TestServerOlderThan(t *testing.T) {
	for _, tt := range []struct {
		server, since string
		want          bool
	}{
		{"", "4.5.0", false},
		{"4.4.0", "", false},
		{"not a version", "4.5.0", false},
		{"4.4.0", "not a version", false},
		{"4.4.0", "4.5.0", true},
		{"4.5.0", "4.5.0", false},
		{"4.7.5", "4.5.0", false},
	} {
		if got := serverOlderThan(tt.server, tt.since); got != tt.want {
			t.Errorf("serverOlderThan(%q, %q) = %v, want %v", tt.server, tt.since, got, tt.want)
		}
	}
}
