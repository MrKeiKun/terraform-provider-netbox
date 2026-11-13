data "netbox_device_front_port" "example" {
  name      = "example"
  device_id = 1
}

output "device_front_port_id" {
  value = data.netbox_device_front_port.example.id
}
