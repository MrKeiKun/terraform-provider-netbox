data "netbox_circuit" "example" {
  id          = 1
  provider_id = 1
}

output "circuit_id" {
  value = data.netbox_circuit.example.id
}
