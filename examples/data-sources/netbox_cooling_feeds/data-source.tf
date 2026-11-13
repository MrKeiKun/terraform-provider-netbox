data "netbox_cooling_feeds" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "cooling_feeds_ids" {
  value = data.netbox_cooling_feeds.example.cooling_feeds[*].id
}
