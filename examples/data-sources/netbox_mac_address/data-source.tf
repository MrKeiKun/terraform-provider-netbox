data "netbox_mac_address" "example" {
  id = 1
}

output "mac_address_id" {
  value = data.netbox_mac_address.example.id
}
