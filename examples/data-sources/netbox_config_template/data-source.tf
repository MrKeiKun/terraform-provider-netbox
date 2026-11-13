data "netbox_config_template" "example" {
  name = "example"
}

output "config_template_id" {
  value = data.netbox_config_template.example.id
}
