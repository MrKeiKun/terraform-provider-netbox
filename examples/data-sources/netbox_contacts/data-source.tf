data "netbox_contacts" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "contacts_ids" {
  value = data.netbox_contacts.example.contacts[*].id
}
