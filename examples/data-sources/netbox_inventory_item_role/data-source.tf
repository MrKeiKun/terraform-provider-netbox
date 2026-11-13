data "netbox_inventory_item_role" "example" {
  id = 1
}

output "inventory_item_role_id" {
  value = data.netbox_inventory_item_role.example.id
}
