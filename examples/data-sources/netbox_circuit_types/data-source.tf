data "netbox_circuit_types" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "circuit_types_ids" {
  value = data.netbox_circuit_types.example.circuit_types[*].id
}
