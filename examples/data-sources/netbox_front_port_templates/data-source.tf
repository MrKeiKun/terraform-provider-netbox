data "netbox_front_port_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "front_port_templates_ids" {
  value = data.netbox_front_port_templates.example.front_port_templates[*].id
}
