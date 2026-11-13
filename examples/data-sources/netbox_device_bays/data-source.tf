data "netbox_device_bays" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "device_bays_ids" {
  value = data.netbox_device_bays.example.device_bays[*].id
}
