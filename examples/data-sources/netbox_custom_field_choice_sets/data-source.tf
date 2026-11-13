data "netbox_custom_field_choice_sets" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "custom_field_choice_sets_ids" {
  value = data.netbox_custom_field_choice_sets.example.custom_field_choice_sets[*].id
}
