resource "netbox_virtual_device_context" "example" {
  name      = "example"
  device_id = netbox_device.example.id
  status    = "active"
}
