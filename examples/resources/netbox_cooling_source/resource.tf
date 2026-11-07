resource "netbox_cooling_source" "example" {
  name = "example"
  site_id = netbox_site.test.id
  type = "chiller"
  status = "active"
  cooling_capacity = 100
}
