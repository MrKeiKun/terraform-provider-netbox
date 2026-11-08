data "netbox_vpn_tunnel_terminations" "all" {}

output "terminations" {
  value = data.netbox_vpn_tunnel_terminations.all.terminations
}
