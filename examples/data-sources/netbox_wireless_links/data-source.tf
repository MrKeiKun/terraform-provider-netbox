data "netbox_wireless_links" "example" {
  filter {
    name  = "status"
    value = "example"
  }
  limit = 10
}

output "wireless_links_ids" {
  value = data.netbox_wireless_links.example.wireless_links[*].id
}
