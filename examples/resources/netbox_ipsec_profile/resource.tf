resource "netbox_ipsec_profile" "example" {
  name            = "example"
  mode            = "esp"
  ike_policy_id   = netbox_ike_policy.example.id
  ipsec_policy_id = netbox_ipsec_policy.example.id
}
