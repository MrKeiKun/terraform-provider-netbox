data "netbox_console_server_port_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "console_server_port_templates_ids" {
  value = data.netbox_console_server_port_templates.example.console_server_port_templates[*].id
}
