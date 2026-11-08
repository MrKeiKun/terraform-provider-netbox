data "netbox_device_module_bays" "example" {
  filter {
    name  = "device_id"
    value = "1"
  }
}

output "device_module_bays" {
  value = data.netbox_device_module_bays.example.module_bays
}
