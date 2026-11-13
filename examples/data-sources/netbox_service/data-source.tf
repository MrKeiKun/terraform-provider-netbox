data "netbox_service" "example" {
  name = "example"
}

output "service_id" {
  value = data.netbox_service.example.id
}
