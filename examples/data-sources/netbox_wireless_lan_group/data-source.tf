data "netbox_wireless_lan_group" "example" {
  id = 1
}

output "wireless_lan_group_id" {
  value = data.netbox_wireless_lan_group.example.id
}
