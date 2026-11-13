data "netbox_cable" "example" {
  id = 1
}

output "cable_id" {
  value = data.netbox_cable.example.id
}
