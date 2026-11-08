resource "netbox_cooling_outflow" "example" {
  name      = "example"
  device_id = netbox_device.example.id
}
