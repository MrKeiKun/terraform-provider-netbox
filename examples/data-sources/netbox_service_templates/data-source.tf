data "netbox_service_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "service_templates_ids" {
  value = data.netbox_service_templates.example.service_templates[*].id
}
