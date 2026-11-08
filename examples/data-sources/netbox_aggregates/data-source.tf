data "netbox_aggregates" "example" {
  filter {
    name  = "family"
    value = "4"
  }
}

output "aggregates" {
  value = data.netbox_aggregates.example.aggregates
}
