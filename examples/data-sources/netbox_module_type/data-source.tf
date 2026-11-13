data "netbox_module_type" "example" {
  id = 1
}

output "module_type_id" {
  value = data.netbox_module_type.example.id
}
