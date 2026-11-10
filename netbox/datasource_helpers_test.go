package netbox

import (
	"reflect"
	"testing"
)

// Upstream issue 919: NetBox allows devices without a name. The JSON reader
// helpers must treat a null or missing name as an empty string, so a name
// filter never panics on such a device.
func TestJStrHandlesUnnamedObjects(t *testing.T) {
	for _, o := range []map[string]interface{}{
		{"name": nil},
		{},
		{"name": 5.0},
	} {
		if got := jStr(o, "name"); got != "" {
			t.Errorf("jStr(%v) = %q, want empty string", o, got)
		}
	}
	if got := jStr(map[string]interface{}{"name": "a"}, "name"); got != "a" {
		t.Errorf("jStr returned %q, want a", got)
	}
}

func TestJTime(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   map[string]interface{}
		want string
	}{
		{"utc", map[string]interface{}{"t": "2024-01-02T03:04:05Z"}, "2024-01-02T03:04:05.000Z"},
		{"offset", map[string]interface{}{"t": "2024-01-02T03:04:05+02:00"}, "2024-01-02T01:04:05.000Z"},
		{"fraction", map[string]interface{}{"t": "2024-01-02T03:04:05.123456Z"}, "2024-01-02T03:04:05.123Z"},
		{"not a time", map[string]interface{}{"t": "yesterday"}, "yesterday"},
		{"missing", map[string]interface{}{}, ""},
		{"null", map[string]interface{}{"t": nil}, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := jTime(tt.in, "t"); got != tt.want {
				t.Errorf("jTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJChoiceMap(t *testing.T) {
	got := jChoiceMap(map[string]interface{}{
		"status": map[string]interface{}{"label": "Active", "value": "active"},
	}, "status")
	want := map[string]string{"label": "Active", "value": "active"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("jChoiceMap() = %v, want %v", got, want)
	}

	// Missing parts of the object are empty strings.
	got = jChoiceMap(map[string]interface{}{"status": map[string]interface{}{}}, "status")
	want = map[string]string{"label": "", "value": ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("jChoiceMap() on empty object = %v, want %v", got, want)
	}

	for _, o := range []map[string]interface{}{{}, {"status": nil}, {"status": "active"}} {
		if got := jChoiceMap(o, "status"); got != nil {
			t.Errorf("jChoiceMap(%v) = %v, want nil", o, got)
		}
	}
}

func TestJRanges(t *testing.T) {
	got := jRanges(map[string]interface{}{
		"ranges": []interface{}{
			[]interface{}{1.0, 10.0},
			[]interface{}{20.0, 30.0, 99.0},
			[]interface{}{5.0},      // too short, skipped
			"not a pair",            // skipped
			[]interface{}{"a", 2.0}, // a start that is not a number is 0
		},
	}, "ranges")
	want := []map[string]int64{
		{"start": 1, "end": 10},
		{"start": 20, "end": 30},
		{"start": 0, "end": 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("jRanges() = %v, want %v", got, want)
	}

	// No ranges gives an empty list, not nil.
	for _, o := range []map[string]interface{}{{}, {"ranges": nil}} {
		got := jRanges(o, "ranges")
		if got == nil || len(got) != 0 {
			t.Errorf("jRanges(%v) = %#v, want an empty list", o, got)
		}
	}
}
