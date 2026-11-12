package fwresource

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCustomFieldValueToString(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input any
		want  string
	}{
		{"null", nil, ""},
		{"string", "value", "value"},
		{"bool", true, "true"},
		{"integer valued float", float64(42), "42"},
		{"float", 3.14, "3.14"},
		{"int", 7, "7"},
		{"int64", int64(100), "100"},
		{"select", map[string]any{"value": "a", "label": "A"}, "a"},
		{"object reference", map[string]any{"id": float64(5), "url": "http://x/5/", "display": "five"}, "5"},
		{"json object stays json", map[string]any{"value": "a", "other": "b"}, `{"other":"b","value":"a"}`},
		{"list of scalars", []any{"a", "b"}, `["a","b"]`},
		{"multi select", []any{map[string]any{"value": "a", "label": "A"}, map[string]any{"value": "b", "label": "B"}}, `["a","b"]`},
		{"multi object", []any{map[string]any{"id": float64(1), "url": "u", "display": "d"}, map[string]any{"id": float64(2), "url": "u", "display": "d"}}, `[1,2]`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := customFieldValueToString(tt.input); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCustomFieldValuesEqual(t *testing.T) {
	for _, tt := range []struct {
		a, b string
		want bool
	}{
		{"a", "a", true},
		{" a ", "a", true},
		{"a", "b", false},
		{`{"a": 1}`, `{"a":1}`, true},
		{`["b","a"]`, `["a","b"]`, true},
		{`[1,2]`, `["1","2"]`, true},
		{`[1,2]`, `[1,3]`, false},
	} {
		if got := customFieldValuesEqual(tt.a, tt.b); got != tt.want {
			t.Errorf("customFieldValuesEqual(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCustomFieldsFromObject(t *testing.T) {
	cf := map[string]any{"managed": "1", "other": "x", "empty": nil}

	// only managed fields are kept when some are configured
	got := customFieldsFromObject(cf, map[string]string{"managed": "1"}, false, true)
	if want := map[string]string{"managed": "1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("managed only: got %v, want %v", got, want)
	}

	// everything that is set is taken over when nothing was configured, e.g. after an import
	got = customFieldsFromObject(cf, nil, true, false)
	if want := map[string]string{"managed": "1", "other": "x"}; !reflect.DeepEqual(got, want) {
		t.Errorf("adopt all: got %v, want %v", got, want)
	}

	// the formatting of the configuration is kept when the content is the same
	got = customFieldsFromObject(map[string]any{"j": map[string]any{"a": float64(1)}}, map[string]string{"j": `{ "a": 1 }`}, false, true)
	if got["j"] != `{ "a": 1 }` {
		t.Errorf("formatting not kept: %v", got)
	}

	// a configured field that NetBox does not report keeps its value
	got = customFieldsFromObject(map[string]any{}, map[string]string{"gone": "v"}, false, true)
	if got["gone"] != "v" {
		t.Errorf("missing field dropped: %v", got)
	}
}

func TestCustomFieldValueToStringEdgeCases(t *testing.T) {
	if got := customFieldValueToString(json.Number("12")); got != "12" {
		t.Errorf("json number = %q", got)
	}
	if got := customFieldValueToString([]any{map[string]any{"value": "a", "label": "A"}, map[string]any{"id": float64(3), "name": "n"}, "plain"}); got != `["a",{"id":3,"name":"n"},"plain"]` {
		t.Errorf("list of choices = %q", got)
	}
	if got := customFieldValueToString([]any{make(chan int)}); got == "" {
		t.Error("an unencodable list gave nothing")
	}
	if got := customFieldValueToString(make(chan int)); got == "" {
		t.Error("an unencodable value gave nothing")
	}
	if !customFieldValuesEqual("1", "1.0") {
		t.Error("1 and 1.0 should be the same number")
	}
}
