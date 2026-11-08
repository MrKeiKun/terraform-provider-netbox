data "netbox_ip_ranges" "example" {
  filter {
    name  = "role"
    value = "lab"
  }
}

output "ip_ranges" {
  value = data.netbox_ip_ranges.example.ip_ranges
}
