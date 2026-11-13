data "netbox_module_bay_template" "example" {
  name = "example"
}

output "module_bay_template_id" {
  value = data.netbox_module_bay_template.example.id
}
