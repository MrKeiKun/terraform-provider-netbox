data "netbox_aggregate" "example" {
  id = 1
}

output "aggregate_id" {
  value = data.netbox_aggregate.example.id
}
