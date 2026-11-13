data "netbox_permissions" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "permissions_ids" {
  value = data.netbox_permissions.example.permissions[*].id
}
