data "netbox_virtual_machine_types" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "virtual_machine_types_ids" {
  value = data.netbox_virtual_machine_types.example.virtual_machine_types[*].id
}
