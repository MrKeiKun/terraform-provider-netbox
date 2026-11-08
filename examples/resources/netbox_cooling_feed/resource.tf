resource "netbox_cooling_feed" "example" {
  name              = "example"
  cooling_source_id = netbox_cooling_source.example.id
}
