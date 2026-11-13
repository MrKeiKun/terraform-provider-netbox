package fwprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEnvBool(t *testing.T) {
	for _, tt := range []struct {
		name    string
		value   types.Bool
		env     string
		want    bool
		wantErr bool
	}{
		{"configured wins", types.BoolValue(true), "false", true, false},
		{"from environment", types.BoolNull(), "true", true, false},
		{"unset uses default", types.BoolNull(), "", true, false},
		{"invalid is an error", types.BoolNull(), "yes please", true, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NETBOX_TEST_BOOL", tt.env)
			var diags diag.Diagnostics
			got := envBool(tt.value, "NETBOX_TEST_BOOL", true, &diags)
			if got != tt.want || diags.HasError() != tt.wantErr {
				t.Fatalf("got %v, error %v; want %v, error %v", got, diags.HasError(), tt.want, tt.wantErr)
			}
		})
	}
}

func TestEnvInt(t *testing.T) {
	for _, tt := range []struct {
		name    string
		value   types.Int64
		env     string
		want    int
		wantErr bool
	}{
		{"configured wins", types.Int64Value(7), "9", 7, false},
		{"from environment", types.Int64Null(), "9", 9, false},
		{"unset uses default", types.Int64Null(), "", 3, false},
		{"zero is allowed", types.Int64Value(0), "", 0, false},
		{"invalid is an error", types.Int64Null(), "10s", 3, true},
		{"negative is an error", types.Int64Value(-1), "", 3, true},
		{"too large is an error", types.Int64Value(1 << 40), "", 3, true},
		{"too large in the environment is an error", types.Int64Null(), "99999999999", 3, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NETBOX_TEST_INT", tt.env)
			var diags diag.Diagnostics
			got := envInt(tt.value, "NETBOX_TEST_INT", 3, &diags)
			if got != tt.want || diags.HasError() != tt.wantErr {
				t.Fatalf("got %v, error %v; want %v, error %v", got, diags.HasError(), tt.want, tt.wantErr)
			}
		})
	}
}
