data "netbox_device_interfaces" "example" {
  filter {
    name  = "device_id"
    value = "1"
  }
}

output "device_interfaces" {
  value = data.netbox_device_interfaces.example.interfaces
}
