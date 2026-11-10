package netbox

import (
	"context"
	"reflect"
	"testing"
)

func TestResourceNetboxVirtualDiskStateUpgradeV0(t *testing.T) {
	for _, tt := range []struct {
		name     string
		state    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "size in GB becomes MB",
			state:    map[string]interface{}{"name": "d", "size_gb": float64(2)},
			expected: map[string]interface{}{"name": "d", "size_mb": float64(2000)},
		},
		{
			name:     "size that is not a number becomes 0",
			state:    map[string]interface{}{"name": "d", "size_gb": "2"},
			expected: map[string]interface{}{"name": "d", "size_mb": float64(0)},
		},
		{
			name:     "missing size becomes 0",
			state:    map[string]interface{}{"name": "d"},
			expected: map[string]interface{}{"name": "d", "size_mb": float64(0)},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resourceNetboxVirtualDiskStateUpgradeV0(context.Background(), tt.state, nil)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestResourceNetboxVirtualDiskResourceV0(t *testing.T) {
	r := resourceNetboxVirtualDiskResourceV0()
	for _, k := range []string{"name", "size_mb", "virtual_machine_id"} {
		if s, ok := r.Schema[k]; !ok || !s.Required {
			t.Errorf("attribute %s should be required", k)
		}
	}
}
