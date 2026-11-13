data "netbox_rack_roles" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "rack_roles_ids" {
  value = data.netbox_rack_roles.example.rack_roles[*].id
}
