data "netbox_power_feed" "example" {
  name    = "example"
  rack_id = 1
}

output "power_feed_id" {
  value = data.netbox_power_feed.example.id
}
