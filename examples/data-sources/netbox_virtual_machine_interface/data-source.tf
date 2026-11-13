data "netbox_virtual_machine_interface" "example" {
  name               = "example"
  virtual_machine_id = 1
}

output "virtual_machine_interface_id" {
  value = data.netbox_virtual_machine_interface.example.id
}
