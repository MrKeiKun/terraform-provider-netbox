data "netbox_rack_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "rack_groups_ids" {
  value = data.netbox_rack_groups.example.rack_groups[*].id
}
