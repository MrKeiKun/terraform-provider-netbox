data "netbox_sites" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "sites_ids" {
  value = data.netbox_sites.example.sites[*].id
}
