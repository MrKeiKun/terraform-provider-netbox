data "netbox_circuit_group_assignment" "example" {
  id       = 1
  group_id = 1
}

output "circuit_group_assignment_id" {
  value = data.netbox_circuit_group_assignment.example.id
}
