data "netbox_vlan_translation_rule" "example" {
  id = 1
}

output "vlan_translation_rule_id" {
  value = data.netbox_vlan_translation_rule.example.id
}
