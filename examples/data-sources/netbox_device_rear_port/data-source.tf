data "netbox_device_rear_port" "example" {
  name      = "example"
  device_id = 1
}

output "device_rear_port_id" {
  value = data.netbox_device_rear_port.example.id
}
