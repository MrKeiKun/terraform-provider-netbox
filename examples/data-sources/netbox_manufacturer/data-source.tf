data "netbox_manufacturer" "example" {
  id = 1
}

output "manufacturer_id" {
  value = data.netbox_manufacturer.example.id
}
