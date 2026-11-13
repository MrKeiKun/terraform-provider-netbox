data "netbox_device_types" "example" {
  filter {
    name  = "slug"
    value = "example"
  }
  limit = 10
}

output "device_types_ids" {
  value = data.netbox_device_types.example.device_types[*].id
}
