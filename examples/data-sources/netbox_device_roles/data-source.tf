data "netbox_device_roles" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "device_roles_ids" {
  value = data.netbox_device_roles.example.device_roles[*].id
}
