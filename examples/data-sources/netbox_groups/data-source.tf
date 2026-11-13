data "netbox_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "groups_ids" {
  value = data.netbox_groups.example.groups[*].id
}
