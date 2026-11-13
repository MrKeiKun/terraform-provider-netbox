data "netbox_power_outlet_template" "example" {
  name = "example"
}

output "power_outlet_template_id" {
  value = data.netbox_power_outlet_template.example.id
}
