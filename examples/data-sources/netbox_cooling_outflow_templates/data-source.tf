data "netbox_cooling_outflow_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "cooling_outflow_templates_ids" {
  value = data.netbox_cooling_outflow_templates.example.cooling_outflow_templates[*].id
}
