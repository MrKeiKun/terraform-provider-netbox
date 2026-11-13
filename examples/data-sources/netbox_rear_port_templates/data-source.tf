data "netbox_rear_port_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "rear_port_templates_ids" {
  value = data.netbox_rear_port_templates.example.rear_port_templates[*].id
}
