package fwprovider

import (
	"testing"
)

func TestAliasResources(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	for _, name := range []string{"netbox_virtual_machine_interface", "netbox_circuit_provider_account", "netbox_available_asn"} {
		if h.schemas.ResourceSchemas[name] == nil {
			t.Errorf("%s is not a resource", name)
		}
	}
	old := h.schemas.ResourceSchemas["netbox_interface"]
	if old == nil || old.Block.Deprecated != true {
		t.Error("netbox_interface is not marked as deprecated")
	}
	if h.schemas.ResourceSchemas["netbox_virtual_machine_interface"].Block.Deprecated {
		t.Error("netbox_virtual_machine_interface is deprecated")
	}
}
