data "netbox_racks" "example" {
  filter {
    name  = "site_id"
    value = "1"
  }
}

output "racks" {
  value = data.netbox_racks.example.racks
}
