data "netbox_device_interface" "example" {
  name      = "example"
  device_id = 1
}

output "device_interface_id" {
  value = data.netbox_device_interface.example.id
}
