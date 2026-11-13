data "netbox_cooling_sources" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "cooling_sources_ids" {
  value = data.netbox_cooling_sources.example.cooling_sources[*].id
}
