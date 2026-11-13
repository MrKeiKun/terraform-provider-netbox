data "netbox_module_bay_types" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "module_bay_types_ids" {
  value = data.netbox_module_bay_types.example.module_bay_types[*].id
}
