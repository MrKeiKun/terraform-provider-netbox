data "netbox_device_front_ports" "example" {
  filter {
    name  = "device_id"
    value = "1"
  }
}

output "device_front_ports" {
  value = data.netbox_device_front_ports.example.front_ports
}
