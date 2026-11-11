package fwprovider

import (
	"context"
	"reflect"
	"testing"
)

func TestUpgradeVirtualDiskV0(t *testing.T) {
	for _, tt := range []struct {
		name     string
		state    map[string]any
		expected map[string]any
	}{
		{"Zero", map[string]any{"size_gb": float64(0)}, map[string]any{"size_mb": float64(0)}},
		{"NonZero", map[string]any{"size_gb": float64(123)}, map[string]any{"size_mb": float64(123000)}},
		{"Invalid", map[string]any{"size_gb": "foo"}, map[string]any{"size_mb": float64(0)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := upgradeVirtualDiskV0(context.Background(), tt.state)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", tt.expected, actual)
			}
		})
	}
}

func TestUpgradeVirtualMachineV1(t *testing.T) {
	for _, tt := range []struct {
		name     string
		state    map[string]any
		expected map[string]any
	}{
		{"Zero", map[string]any{"disk_size_gb": float64(0)}, map[string]any{"disk_size_mb": float64(0)}},
		{"NonZero", map[string]any{"disk_size_gb": float64(123)}, map[string]any{"disk_size_mb": float64(123000)}},
		{"Invalid", map[string]any{"disk_size_gb": "foo"}, map[string]any{"disk_size_mb": float64(0)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := upgradeVirtualMachineV1(context.Background(), tt.state)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", tt.expected, actual)
			}
		})
	}
}

func TestUpgradeVirtualMachineV0(t *testing.T) {
	for _, tt := range []struct {
		name     string
		state    map[string]any
		expected map[string]any
	}{
		{"StringVcpus", map[string]any{"vcpus": "2.5", "disk_size_gb": float64(1)}, map[string]any{"vcpus": 2.5, "disk_size_mb": float64(1000)}},
		{"InvalidVcpus", map[string]any{"vcpus": "x", "disk_size_gb": float64(0)}, map[string]any{"vcpus": float64(0), "disk_size_mb": float64(0)}},
		{"NumericVcpus", map[string]any{"vcpus": float64(4), "disk_size_gb": float64(0)}, map[string]any{"vcpus": float64(4), "disk_size_mb": float64(0)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := upgradeVirtualMachineV0(context.Background(), tt.state)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", tt.expected, actual)
			}
		})
	}
}
