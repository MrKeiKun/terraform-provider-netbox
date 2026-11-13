data "netbox_contact_assignments" "example" {
  filter {
    name  = "contact_id"
    value = "example"
  }
  limit = 10
}

output "contact_assignments_ids" {
  value = data.netbox_contact_assignments.example.contact_assignments[*].id
}
