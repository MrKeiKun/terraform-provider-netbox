data "netbox_vpn_tunnels" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "vpn_tunnels_ids" {
  value = data.netbox_vpn_tunnels.example.tunnels[*].id
}
