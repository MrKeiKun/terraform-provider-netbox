data "netbox_power_feeds" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "power_feeds_ids" {
  value = data.netbox_power_feeds.example.power_feeds[*].id
}
