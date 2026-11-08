data "netbox_manufacturers" "example" {
  filter {
    name  = "name"
    value = "Juniper"
  }
}

output "manufacturers" {
  value = data.netbox_manufacturers.example.manufacturers
}
