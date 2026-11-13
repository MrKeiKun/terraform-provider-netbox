data "netbox_device_bay" "example" {
  name      = "example"
  device_id = 1
}

output "device_bay_id" {
  value = data.netbox_device_bay.example.id
}
