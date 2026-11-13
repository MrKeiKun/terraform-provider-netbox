data "netbox_module_bay_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "module_bay_templates_ids" {
  value = data.netbox_module_bay_templates.example.module_bay_templates[*].id
}
