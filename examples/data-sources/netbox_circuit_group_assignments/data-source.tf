data "netbox_circuit_group_assignments" "example" {
  filter {
    name  = "circuit"
    value = "example"
  }
  limit = 10
}

output "circuit_group_assignments_ids" {
  value = data.netbox_circuit_group_assignments.example.circuit_group_assignments[*].id
}
