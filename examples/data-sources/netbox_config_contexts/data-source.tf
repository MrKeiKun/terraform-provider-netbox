data "netbox_config_contexts" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "config_contexts_ids" {
  value = data.netbox_config_contexts.example.config_contexts[*].id
}
