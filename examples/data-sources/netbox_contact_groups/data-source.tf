data "netbox_contact_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "contact_groups_ids" {
  value = data.netbox_contact_groups.example.contact_groups[*].id
}
