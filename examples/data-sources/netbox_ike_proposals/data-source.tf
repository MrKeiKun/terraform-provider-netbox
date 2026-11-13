data "netbox_ike_proposals" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "ike_proposals_ids" {
  value = data.netbox_ike_proposals.example.ike_proposals[*].id
}
