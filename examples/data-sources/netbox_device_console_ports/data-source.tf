data "netbox_device_console_ports" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "device_console_ports_ids" {
  value = data.netbox_device_console_ports.example.device_console_ports[*].id
}
