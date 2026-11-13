data "netbox_virtual_circuit_termination" "example" {
  id = 1
}

output "virtual_circuit_termination_id" {
  value = data.netbox_virtual_circuit_termination.example.id
}
