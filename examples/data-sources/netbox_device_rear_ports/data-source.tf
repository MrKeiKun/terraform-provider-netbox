data "netbox_device_rear_ports" "example" {
  filter {
    name  = "device_id"
    value = "1"
  }
}

output "device_rear_ports" {
  value = data.netbox_device_rear_ports.example.rear_ports
}
