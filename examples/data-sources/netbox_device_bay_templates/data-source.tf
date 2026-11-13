data "netbox_device_bay_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "device_bay_templates_ids" {
  value = data.netbox_device_bay_templates.example.device_bay_templates[*].id
}
