resource "netbox_virtual_machine_type" "example" {
  name = "example"
  default_vcpus = 2
  default_memory = 2048
}
