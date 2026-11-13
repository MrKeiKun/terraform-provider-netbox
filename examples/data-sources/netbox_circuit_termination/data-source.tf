data "netbox_circuit_termination" "example" {
  id         = 1
  circuit_id = 1
}

output "circuit_termination_id" {
  value = data.netbox_circuit_termination.example.id
}
