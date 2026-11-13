data "netbox_device_module_bay" "example" {
  name      = "example"
  device_id = 1
}

output "device_module_bay_id" {
  value = data.netbox_device_module_bay.example.id
}
