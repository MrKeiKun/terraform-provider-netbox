data "netbox_cable_bundles" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "cable_bundles_ids" {
  value = data.netbox_cable_bundles.example.cable_bundles[*].id
}
