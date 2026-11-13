data "netbox_virtual_machine" "example" {
  name      = "example"
  device_id = 1
}

output "virtual_machine_id" {
  value = data.netbox_virtual_machine.example.id
}
