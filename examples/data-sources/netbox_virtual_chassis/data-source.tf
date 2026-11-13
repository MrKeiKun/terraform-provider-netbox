data "netbox_virtual_chassis" "example" {
  name = "example"
}

output "virtual_chassis_id" {
  value = data.netbox_virtual_chassis.example.id
}
