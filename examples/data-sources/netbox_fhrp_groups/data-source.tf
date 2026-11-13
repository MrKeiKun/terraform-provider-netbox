data "netbox_fhrp_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "fhrp_groups_ids" {
  value = data.netbox_fhrp_groups.example.fhrp_groups[*].id
}
