data "netbox_device_console_server_ports" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "device_console_server_ports_ids" {
  value = data.netbox_device_console_server_ports.example.device_console_server_ports[*].id
}
