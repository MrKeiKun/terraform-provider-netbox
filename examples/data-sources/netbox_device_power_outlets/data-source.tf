data "netbox_device_power_outlets" "example" {
  filter {
    name  = "device_id"
    value = "1"
  }
}

output "device_power_outlets" {
  value = data.netbox_device_power_outlets.example.power_outlets
}
