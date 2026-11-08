resource "netbox_cooling_outflow_template" "example" {
  name           = "example"
  device_type_id = netbox_device_type.example.id
}
