data "netbox_custom_field" "example" {
  name = "example"
}

output "custom_field_id" {
  value = data.netbox_custom_field.example.id
}
