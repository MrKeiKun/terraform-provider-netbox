data "netbox_rear_port_template" "example" {
  name = "example"
}

output "rear_port_template_id" {
  value = data.netbox_rear_port_template.example.id
}
