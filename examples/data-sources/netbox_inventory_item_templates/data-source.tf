data "netbox_inventory_item_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "inventory_item_templates_ids" {
  value = data.netbox_inventory_item_templates.example.inventory_item_templates[*].id
}
