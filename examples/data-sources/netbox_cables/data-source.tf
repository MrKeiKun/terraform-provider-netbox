data "netbox_cables" "example" {
  filter {
    name  = "status"
    value = "connected"
  }
}

output "cables" {
  value = data.netbox_cables.example.cables
}
