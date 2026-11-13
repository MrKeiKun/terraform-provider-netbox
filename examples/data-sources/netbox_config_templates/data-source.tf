data "netbox_config_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "config_templates_ids" {
  value = data.netbox_config_templates.example.config_templates[*].id
}
