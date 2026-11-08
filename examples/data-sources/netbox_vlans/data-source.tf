data "netbox_vlans" "example" {
  filter {
    name  = "vid"
    value = "100"
  }
}

output "vlans" {
  value = data.netbox_vlans.example.vlans
}
