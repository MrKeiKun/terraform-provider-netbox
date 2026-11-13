data "netbox_circuit_terminations" "example" {
  filter {
    name  = "cable_connector"
    value = "example"
  }
  limit = 10
}

output "circuit_terminations_ids" {
  value = data.netbox_circuit_terminations.example.circuit_terminations[*].id
}
