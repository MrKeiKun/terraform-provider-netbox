data "netbox_contact_assignment" "example" {
  id = 1
}

output "contact_assignment_id" {
  value = data.netbox_contact_assignment.example.id
}
