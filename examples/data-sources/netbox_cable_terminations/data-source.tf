data "netbox_cable_terminations" "example" {
  filter {
    name  = "cable"
    value = "example"
  }
  limit = 10
}

output "cable_terminations_ids" {
  value = data.netbox_cable_terminations.example.cable_terminations[*].id
}
