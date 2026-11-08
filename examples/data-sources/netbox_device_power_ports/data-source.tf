data "netbox_device_power_ports" "example" {
  filter {
    name  = "device_id"
    value = "1"
  }
}

output "device_power_ports" {
  value = data.netbox_device_power_ports.example.power_ports
}
