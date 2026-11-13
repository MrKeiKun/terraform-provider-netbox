data "netbox_custom_fields" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "custom_fields_ids" {
  value = data.netbox_custom_fields.example.custom_fields[*].id
}
