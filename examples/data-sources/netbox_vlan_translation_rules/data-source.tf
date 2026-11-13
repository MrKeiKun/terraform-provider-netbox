data "netbox_vlan_translation_rules" "example" {
  filter {
    name  = "created"
    value = "example"
  }
  limit = 10
}

output "vlan_translation_rules_ids" {
  value = data.netbox_vlan_translation_rules.example.vlan_translation_rules[*].id
}
