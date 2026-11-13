data "netbox_wireless_lan_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "wireless_lan_groups_ids" {
  value = data.netbox_wireless_lan_groups.example.wireless_lan_groups[*].id
}
