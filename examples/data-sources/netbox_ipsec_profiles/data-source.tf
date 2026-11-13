data "netbox_ipsec_profiles" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "ipsec_profiles_ids" {
  value = data.netbox_ipsec_profiles.example.ipsec_profiles[*].id
}
