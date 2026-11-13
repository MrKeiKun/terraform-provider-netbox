data "netbox_circuits" "example" {
  filter {
    name  = "status"
    value = "example"
  }
  limit = 10
}

output "circuits_ids" {
  value = data.netbox_circuits.example.circuits[*].id
}
