resource "netbox_notification" "example" {
  object_type = "dcim.site"
  object_id   = netbox_site.example.id
  user_id     = netbox_user.example.id
  event_type  = "object_updated"
}
