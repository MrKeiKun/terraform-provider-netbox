package netbox

import "testing"

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
