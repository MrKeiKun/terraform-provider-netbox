data "netbox_virtual_circuit_types" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "virtual_circuit_types_ids" {
  value = data.netbox_virtual_circuit_types.example.virtual_circuit_types[*].id
}
