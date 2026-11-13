data "netbox_cooling_intakes" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "cooling_intakes_ids" {
  value = data.netbox_cooling_intakes.example.cooling_intakes[*].id
}
