data "netbox_inventory_item" "example" {
  name      = "example"
  device_id = 1
}

output "inventory_item_id" {
  value = data.netbox_inventory_item.example.id
}
