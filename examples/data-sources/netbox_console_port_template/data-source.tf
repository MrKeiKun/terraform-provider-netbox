data "netbox_console_port_template" "example" {
  name = "example"
}

output "console_port_template_id" {
  value = data.netbox_console_port_template.example.id
}
