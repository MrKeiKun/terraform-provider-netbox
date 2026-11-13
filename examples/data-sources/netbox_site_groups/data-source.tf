data "netbox_site_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "site_groups_ids" {
  value = data.netbox_site_groups.example.site_groups[*].id
}
