data "netbox_inventory_item_roles" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "inventory_item_roles_ids" {
  value = data.netbox_inventory_item_roles.example.inventory_item_roles[*].id
}
