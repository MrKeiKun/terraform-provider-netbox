data "netbox_ipsec_proposals" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "ipsec_proposals_ids" {
  value = data.netbox_ipsec_proposals.example.ipsec_proposals[*].id
}
