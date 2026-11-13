data "netbox_saved_filters" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "saved_filters_ids" {
  value = data.netbox_saved_filters.example.saved_filters[*].id
}
