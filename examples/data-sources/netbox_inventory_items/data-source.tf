data "netbox_inventory_items" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "inventory_items_ids" {
  value = data.netbox_inventory_items.example.inventory_items[*].id
}
