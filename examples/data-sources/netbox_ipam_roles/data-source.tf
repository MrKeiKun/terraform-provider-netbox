data "netbox_ipam_roles" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "ipam_roles_ids" {
  value = data.netbox_ipam_roles.example.ipam_roles[*].id
}
