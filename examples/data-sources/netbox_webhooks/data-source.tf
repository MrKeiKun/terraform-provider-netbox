data "netbox_webhooks" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "webhooks_ids" {
  value = data.netbox_webhooks.example.webhooks[*].id
}
