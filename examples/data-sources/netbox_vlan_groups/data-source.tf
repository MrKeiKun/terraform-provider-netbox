data "netbox_vlan_groups" "example" {
  filter {
    name  = "name"
    value = "campus"
  }
}

output "vlan_groups" {
  value = data.netbox_vlan_groups.example.vlan_groups
}
