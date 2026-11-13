data "netbox_module_type_profiles" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "module_type_profiles_ids" {
  value = data.netbox_module_type_profiles.example.module_type_profiles[*].id
}
