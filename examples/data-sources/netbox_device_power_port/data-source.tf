data "netbox_device_power_port" "example" {
  name      = "example"
  device_id = 1
}

output "device_power_port_id" {
  value = data.netbox_device_power_port.example.id
}
