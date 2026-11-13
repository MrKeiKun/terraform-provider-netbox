data "netbox_route_targets" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "route_targets_ids" {
  value = data.netbox_route_targets.example.route_targets[*].id
}
