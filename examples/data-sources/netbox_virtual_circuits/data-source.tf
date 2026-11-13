data "netbox_virtual_circuits" "example" {
  filter {
    name  = "status"
    value = "example"
  }
  limit = 10
}

output "virtual_circuits_ids" {
  value = data.netbox_virtual_circuits.example.virtual_circuits[*].id
}
