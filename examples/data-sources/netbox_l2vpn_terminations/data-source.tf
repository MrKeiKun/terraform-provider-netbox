data "netbox_l2vpn_terminations" "example" {
  filter {
    name  = "assigned_object_id"
    value = "example"
  }
  limit = 10
}

output "l2vpn_terminations_ids" {
  value = data.netbox_l2vpn_terminations.example.l2vpn_terminations[*].id
}
