data "netbox_owner_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "owner_groups_ids" {
  value = data.netbox_owner_groups.example.owner_groups[*].id
}
