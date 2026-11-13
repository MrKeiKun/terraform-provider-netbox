data "netbox_power_outlet_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "power_outlet_templates_ids" {
  value = data.netbox_power_outlet_templates.example.power_outlet_templates[*].id
}
