data "netbox_data_sources" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "data_sources_ids" {
  value = data.netbox_data_sources.example.data_sources[*].id
}
