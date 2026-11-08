data "netbox_ip_addresses" "example" {
  filter {
    name  = "device_id"
    value = "1"
  }
}

output "ip_addresses" {
  value = data.netbox_ip_addresses.example.ip_addresses
}
