data "netbox_region" "example" {
  filter {
    name  = "name"
    value = "Europe"
  }
}
