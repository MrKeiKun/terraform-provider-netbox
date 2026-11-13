data "netbox_virtual_machine_interfaces" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "virtual_machine_interfaces_ids" {
  value = data.netbox_virtual_machine_interfaces.example.interfaces[*].id
}
