data "netbox_vpn_tunnel_termination" "example" {
  id = 1
}

output "vpn_tunnel_termination_id" {
  value = data.netbox_vpn_tunnel_termination.example.id
}
