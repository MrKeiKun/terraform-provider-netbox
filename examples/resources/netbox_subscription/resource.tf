resource "netbox_subscription" "example" {
  object_type = "dcim.site"
  object_id   = netbox_site.example.id
  user_id     = netbox_user.example.id
}
