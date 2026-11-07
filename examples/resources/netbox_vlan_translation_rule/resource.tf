resource "netbox_vlan_translation_rule" "example" {
  policy_id  = netbox_vlan_translation_policy.example.id
  local_vid  = 100
  remote_vid = 200
}
