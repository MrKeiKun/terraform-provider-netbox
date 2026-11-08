data "netbox_tunnels" "example" {
  filter {
    name  = "status"
    value = "active"
  }
}

output "tunnels" {
  value = data.netbox_tunnels.example.tunnels
}
