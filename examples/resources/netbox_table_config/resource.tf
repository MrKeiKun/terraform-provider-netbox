resource "netbox_table_config" "example" {
  name        = "Sites with region"
  object_type = "dcim.site"
  table       = "SiteTable"
  columns     = ["name", "status", "region", "tenant"]
  ordering    = ["name"]
  shared      = true
}
