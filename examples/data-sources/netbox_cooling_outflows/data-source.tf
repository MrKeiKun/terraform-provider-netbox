data "netbox_cooling_outflows" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "cooling_outflows_ids" {
  value = data.netbox_cooling_outflows.example.cooling_outflows[*].id
}
