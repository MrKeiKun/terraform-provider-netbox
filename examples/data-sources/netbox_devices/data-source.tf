data "netbox_devices" "example" {
  filter {
    name  = "name"
    value = "router-01"
  }
}

output "devices" {
  value = data.netbox_devices.example.devices
}
