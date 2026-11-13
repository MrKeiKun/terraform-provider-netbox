data "netbox_wireless_lans" "example" {
  filter {
    name  = "status"
    value = "example"
  }
  limit = 10
}

output "wireless_lans_ids" {
  value = data.netbox_wireless_lans.example.wireless_lans[*].id
}
