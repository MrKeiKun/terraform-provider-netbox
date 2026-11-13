data "netbox_cable_termination" "example" {
  id = 1
}

output "cable_termination_id" {
  value = data.netbox_cable_termination.example.id
}
