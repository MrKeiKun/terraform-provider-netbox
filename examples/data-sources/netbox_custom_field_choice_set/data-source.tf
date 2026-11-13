data "netbox_custom_field_choice_set" "example" {
  name = "example"
}

output "custom_field_choice_set_id" {
  value = data.netbox_custom_field_choice_set.example.id
}
