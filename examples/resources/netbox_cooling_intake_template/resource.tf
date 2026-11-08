resource "netbox_cooling_intake_template" "example" {
  name           = "example"
  device_type_id = netbox_device_type.example.id
}
