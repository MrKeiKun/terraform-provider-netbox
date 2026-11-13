data "netbox_device_power_outlet" "example" {
  name      = "example"
  device_id = 1
}

output "device_power_outlet_id" {
  value = data.netbox_device_power_outlet.example.id
}
