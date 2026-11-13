data "netbox_device_console_server_port" "example" {
  name      = "example"
  device_id = 1
}

output "device_console_server_port_id" {
  value = data.netbox_device_console_server_port.example.id
}
