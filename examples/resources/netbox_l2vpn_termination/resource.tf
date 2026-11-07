resource "netbox_l2vpn_termination" "example" {
  l2vpn_id             = netbox_l2vpn.example.id
  assigned_object_type = "ipam.vlan"
  assigned_object_id   = netbox_vlan.example.id
}
