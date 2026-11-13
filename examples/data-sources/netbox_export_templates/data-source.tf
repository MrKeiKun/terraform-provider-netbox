data "netbox_export_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "export_templates_ids" {
  value = data.netbox_export_templates.example.export_templates[*].id
}
