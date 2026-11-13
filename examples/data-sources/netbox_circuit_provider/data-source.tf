data "netbox_circuit_provider" "example" {
  id = 1
}

output "circuit_provider_id" {
  value = data.netbox_circuit_provider.example.id
}
