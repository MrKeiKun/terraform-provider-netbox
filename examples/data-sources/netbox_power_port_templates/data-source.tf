data "netbox_power_port_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "power_port_templates_ids" {
  value = data.netbox_power_port_templates.example.power_port_templates[*].id
}
