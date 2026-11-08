data "netbox_locations" "example" {
  filter {
    name  = "site_id"
    value = "1"
  }
}

output "locations" {
  value = data.netbox_locations.example.locations
}
