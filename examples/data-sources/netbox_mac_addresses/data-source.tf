data "netbox_mac_addresses" "example" {
  filter {
    name  = "assigned"
    value = "example"
  }
  limit = 10
}

output "mac_addresses_ids" {
  value = data.netbox_mac_addresses.example.mac_addresses[*].id
}
