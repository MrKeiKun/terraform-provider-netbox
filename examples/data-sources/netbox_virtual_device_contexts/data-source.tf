data "netbox_virtual_device_contexts" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "virtual_device_contexts_ids" {
  value = data.netbox_virtual_device_contexts.example.virtual_device_contexts[*].id
}
