data "netbox_interface_template" "example" {
  name = "example"
}

output "interface_template_id" {
  value = data.netbox_interface_template.example.id
}
