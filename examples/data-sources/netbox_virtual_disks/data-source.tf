data "netbox_virtual_disks" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "virtual_disks_ids" {
  value = data.netbox_virtual_disks.example.virtual_disks[*].id
}
