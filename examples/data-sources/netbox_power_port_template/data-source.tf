data "netbox_power_port_template" "example" {
  name = "example"
}

output "power_port_template_id" {
  value = data.netbox_power_port_template.example.id
}
