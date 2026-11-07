resource "netbox_inventory_item_template" "example" {
  name           = "example"
  device_type_id = netbox_device_type.example.id
}
