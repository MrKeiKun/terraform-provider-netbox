data "netbox_vpn_tunnel_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "vpn_tunnel_groups_ids" {
  value = data.netbox_vpn_tunnel_groups.example.vpn_tunnel_groups[*].id
}
