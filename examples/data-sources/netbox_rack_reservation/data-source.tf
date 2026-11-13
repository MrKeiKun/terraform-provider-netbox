data "netbox_rack_reservation" "example" {
  id      = 1
  rack_id = 1
}

output "rack_reservation_id" {
  value = data.netbox_rack_reservation.example.id
}
