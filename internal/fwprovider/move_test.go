package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// A moved block turns a resource into the one that has a second name, and keeps what the state holds.
func TestMoveResourceState(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)

	for _, tt := range []struct {
		from, to, state, attr, want string
	}{
		{"netbox_interface", "netbox_virtual_machine_interface", `{"id":"14","name":"eth0","virtual_machine_id":2,"mtu":1500}`, "name", "eth0"},
		{"netbox_provider_account", "netbox_circuit_provider_account", `{"id":"3","account":"a-1","provider_id":1}`, "account", "a-1"},
	} {
		resp, err := h.server.MoveResourceState(context.Background(), &tfprotov6.MoveResourceStateRequest{
			SourceProviderAddress: "registry.terraform.io/mrkeikun/netbox",
			SourceTypeName:        tt.from,
			SourceSchemaVersion:   0,
			SourceState:           &tfprotov6.RawState{JSON: []byte(tt.state)},
			TargetTypeName:        tt.to,
		})
		if err != nil {
			t.Fatalf("%s: %v", tt.from, err)
		}
		h.noErrors(tt.from, resp.Diagnostics)
		if resp.TargetState == nil {
			t.Fatalf("%s: no state", tt.from)
		}
		got := h.decode(h.schemas.ResourceSchemas[tt.to].ValueType(), resp.TargetState)
		if stringValue := func() string { var s string; _ = got[tt.attr].As(&s); return s }(); stringValue != tt.want {
			t.Errorf("%s: %s = %q, want %q", tt.from, tt.attr, stringValue, tt.want)
		}
		var id string
		if err := got["id"].As(&id); err != nil || id == "" {
			t.Errorf("%s: the id was lost: %v", tt.from, got["id"])
		}
	}

	// a resource that is not listed cannot be moved to
	resp, err := h.server.MoveResourceState(context.Background(), &tfprotov6.MoveResourceStateRequest{
		SourceProviderAddress: "registry.terraform.io/mrkeikun/netbox",
		SourceTypeName:        "netbox_site",
		SourceState:           &tfprotov6.RawState{JSON: []byte(`{"id":"1","name":"s","slug":"s"}`)},
		TargetTypeName:        "netbox_virtual_machine_interface",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !h.hasError(resp.Diagnostics) {
		t.Error("a site was moved to a virtual machine interface")
	}
}
