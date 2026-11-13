data "netbox_wireless_lan" "example" {
  id       = 1
  group_id = 1
}

output "wireless_lan_id" {
  value = data.netbox_wireless_lan.example.id
}
