data "netbox_regions" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "regions_ids" {
  value = data.netbox_regions.example.regions[*].id
}
