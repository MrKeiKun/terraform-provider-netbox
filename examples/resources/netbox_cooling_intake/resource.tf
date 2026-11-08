resource "netbox_cooling_intake" "example" {
  name      = "example"
  device_id = netbox_device.example.id
}
