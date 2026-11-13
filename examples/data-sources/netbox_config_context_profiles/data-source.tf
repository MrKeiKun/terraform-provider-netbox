data "netbox_config_context_profiles" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "config_context_profiles_ids" {
  value = data.netbox_config_context_profiles.example.config_context_profiles[*].id
}
