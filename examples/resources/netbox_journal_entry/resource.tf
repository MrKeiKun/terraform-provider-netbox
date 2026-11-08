resource "netbox_journal_entry" "example" {
  assigned_object_type = "dcim.site"
  assigned_object_id   = netbox_site.example.id
  comments             = "example"
}
