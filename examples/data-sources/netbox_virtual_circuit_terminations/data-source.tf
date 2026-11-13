data "netbox_virtual_circuit_terminations" "example" {
  filter {
    name  = "created"
    value = "example"
  }
  limit = 10
}

output "virtual_circuit_terminations_ids" {
  value = data.netbox_virtual_circuit_terminations.example.virtual_circuit_terminations[*].id
}
