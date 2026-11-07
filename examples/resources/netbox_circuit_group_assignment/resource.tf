resource "netbox_circuit_group_assignment" "example" {
  group_id    = netbox_circuit_group.example.id
  member_type = "circuits.circuit"
  member_id   = netbox_circuit.example.id
}
