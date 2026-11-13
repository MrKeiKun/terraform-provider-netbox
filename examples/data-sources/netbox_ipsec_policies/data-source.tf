data "netbox_ipsec_policies" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "ipsec_policies_ids" {
  value = data.netbox_ipsec_policies.example.ipsec_policies[*].id
}
