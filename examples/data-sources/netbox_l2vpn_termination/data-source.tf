data "netbox_l2vpn_termination" "example" {
  id       = 1
  l2vpn_id = 1
}

output "l2vpn_termination_id" {
  value = data.netbox_l2vpn_termination.example.id
}
