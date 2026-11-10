package netbox

import "testing"

func TestCustomFieldValuesEqualLists(t *testing.T) {
	for _, tt := range []struct {
		name string
		a, b string
		want bool
	}{
		{"same text", "abc", "abc", true},
		{"surrounding space", " abc ", "abc", true},
		{"different text", "abc", "abd", false},
		{"json formatting", `{"a": 1,  "b": 2}`, `{"b":2,"a":1}`, true},
		{"list order", `["a","b"]`, `["b","a"]`, true},
		{"number and string ids", `[1,2]`, `["2","1"]`, true},
		{"different lists", `[1,2]`, `[1,3]`, false},
		{"lists of objects keep their order", `[{"id":1},{"id":2}]`, `[{"id":2},{"id":1}]`, false},
		{"lists of objects", `[{"id":1},{"id":2}]`, `[{"id":1}, {"id":2}]`, true},
		{"list in a list", `[[1],[2]]`, `[[1],[2]]`, true},
		{"json and plain text", `[1]`, `one`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := customFieldValuesEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("customFieldValuesEqual(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
