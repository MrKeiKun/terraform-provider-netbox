resource "netbox_virtual_circuit" "example" {
  cid                 = "example"
  provider_network_id = netbox_circuit_provider_network.example.id
  type_id             = netbox_virtual_circuit_type.example.id
}
