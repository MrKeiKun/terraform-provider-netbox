data "netbox_circuit_providers" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "circuit_providers_ids" {
  value = data.netbox_circuit_providers.example.circuit_providers[*].id
}
