data "netbox_vlan_translation_policies" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "vlan_translation_policies_ids" {
  value = data.netbox_vlan_translation_policies.example.vlan_translation_policies[*].id
}
