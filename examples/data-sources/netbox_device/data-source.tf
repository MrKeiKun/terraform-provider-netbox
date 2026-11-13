data "netbox_device" "example" {
  name    = "example"
  rack_id = 1
}

output "device_id" {
  value = data.netbox_device.example.id
}
