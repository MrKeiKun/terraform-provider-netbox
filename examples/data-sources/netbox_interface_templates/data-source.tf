data "netbox_interface_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "interface_templates_ids" {
  value = data.netbox_interface_templates.example.interface_templates[*].id
}
