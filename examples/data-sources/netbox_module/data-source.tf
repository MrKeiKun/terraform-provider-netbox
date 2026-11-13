data "netbox_module" "example" {
  id        = 1
  device_id = 1
}

output "module_id" {
  value = data.netbox_module.example.id
}
