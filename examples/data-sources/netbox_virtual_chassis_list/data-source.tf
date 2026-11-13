data "netbox_virtual_chassis_list" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "virtual_chassis_list_ids" {
  value = data.netbox_virtual_chassis_list.example.virtual_chassis_list[*].id
}
