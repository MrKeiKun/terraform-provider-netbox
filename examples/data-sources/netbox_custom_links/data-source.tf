data "netbox_custom_links" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "custom_links_ids" {
  value = data.netbox_custom_links.example.custom_links[*].id
}
