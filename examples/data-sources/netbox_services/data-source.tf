data "netbox_services" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "services_ids" {
  value = data.netbox_services.example.services[*].id
}
