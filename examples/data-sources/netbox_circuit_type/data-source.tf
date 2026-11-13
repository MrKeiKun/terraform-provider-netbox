data "netbox_circuit_type" "example" {
  id = 1
}

output "circuit_type_id" {
  value = data.netbox_circuit_type.example.id
}
