data "netbox_ike_policies" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "ike_policies_ids" {
  value = data.netbox_ike_policies.example.ike_policies[*].id
}
