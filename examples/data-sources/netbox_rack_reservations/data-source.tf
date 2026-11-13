data "netbox_rack_reservations" "example" {
  filter {
    name  = "status"
    value = "example"
  }
  limit = 10
}

output "rack_reservations_ids" {
  value = data.netbox_rack_reservations.example.rack_reservations[*].id
}
