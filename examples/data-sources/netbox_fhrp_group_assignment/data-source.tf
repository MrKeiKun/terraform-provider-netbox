data "netbox_fhrp_group_assignment" "example" {
  id       = 1
  group_id = 1
}

output "fhrp_group_assignment_id" {
  value = data.netbox_fhrp_group_assignment.example.id
}
