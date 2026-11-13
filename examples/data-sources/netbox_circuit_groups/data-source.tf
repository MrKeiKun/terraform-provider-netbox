data "netbox_circuit_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "circuit_groups_ids" {
  value = data.netbox_circuit_groups.example.circuit_groups[*].id
}
