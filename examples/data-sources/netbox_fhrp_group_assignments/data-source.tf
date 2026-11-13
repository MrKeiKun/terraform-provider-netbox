data "netbox_fhrp_group_assignments" "example" {
  filter {
    name  = "created"
    value = "example"
  }
  limit = 10
}

output "fhrp_group_assignments_ids" {
  value = data.netbox_fhrp_group_assignments.example.fhrp_group_assignments[*].id
}
