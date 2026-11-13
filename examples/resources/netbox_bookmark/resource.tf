resource "netbox_site" "example" {
  name = "Example site"
  slug = "example-site"
}

resource "netbox_bookmark" "example" {
  object_type = "dcim.site"
  object_id   = netbox_site.example.id
  user_id     = netbox_user.example.id
}
