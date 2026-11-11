package fwprovider

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// The server hands the old state to the upgrader of the schema version it was written with.
func TestUpgradeResourceState(t *testing.T) {
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatal(err)
	}

	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		typeName string
		version  int64
		state    string
		field    string
		want     float64
	}{
		{"netbox_virtual_disk", 0, `{"id":"1","name":"d","size_gb":2,"virtual_machine_id":3}`, "size_mb", 2000},
		{"netbox_virtual_machine", 0, `{"id":"1","name":"vm","vcpus":"2","disk_size_gb":1}`, "disk_size_mb", 1000},
		{"netbox_virtual_machine", 1, `{"id":"1","name":"vm","disk_size_gb":3}`, "disk_size_mb", 3000},
	} {
		resp, err := server.UpgradeResourceState(context.Background(), &tfprotov6.UpgradeResourceStateRequest{
			TypeName: tt.typeName,
			Version:  tt.version,
			RawState: &tfprotov6.RawState{JSON: []byte(tt.state)},
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range resp.Diagnostics {
			t.Errorf("%s v%d: %s: %s", tt.typeName, tt.version, d.Summary, d.Detail)
		}
		if resp.UpgradedState == nil {
			t.Fatalf("%s v%d: no upgraded state", tt.typeName, tt.version)
		}
		value, err := resp.UpgradedState.Unmarshal(schemas.ResourceSchemas[tt.typeName].ValueType())
		if err != nil {
			t.Fatal(err)
		}
		var attrs map[string]tftypes.Value
		if err := value.As(&attrs); err != nil {
			t.Fatal(err)
		}
		var got big.Float
		if err := attrs[tt.field].As(&got); err != nil {
			t.Fatal(err)
		}
		if f, _ := got.Float64(); f != tt.want {
			t.Errorf("%s v%d: %s = %v, want %v", tt.typeName, tt.version, tt.field, f, tt.want)
		}
		if _, ok := attrs["size_gb"]; ok {
			t.Errorf("%s v%d: size_gb was not removed", tt.typeName, tt.version)
		}
	}
}
