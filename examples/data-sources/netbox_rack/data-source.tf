data "netbox_rack" "example" {
  name    = "example"
  site_id = 1
}

output "rack_id" {
  value = data.netbox_rack.example.id
}
