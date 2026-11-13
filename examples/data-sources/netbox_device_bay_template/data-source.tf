data "netbox_device_bay_template" "example" {
  name = "example"
}

output "device_bay_template_id" {
  value = data.netbox_device_bay_template.example.id
}
