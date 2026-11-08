resource "netbox_front_port_template" "example" {
  name           = "example"
  device_type_id = netbox_device_type.example.id
  type           = "8p8c"
}
