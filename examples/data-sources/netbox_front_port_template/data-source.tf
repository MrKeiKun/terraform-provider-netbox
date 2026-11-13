data "netbox_front_port_template" "example" {
  name = "example"
}

output "front_port_template_id" {
  value = data.netbox_front_port_template.example.id
}
