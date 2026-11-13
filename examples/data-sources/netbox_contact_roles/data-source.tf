data "netbox_contact_roles" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "contact_roles_ids" {
  value = data.netbox_contact_roles.example.contact_roles[*].id
}
