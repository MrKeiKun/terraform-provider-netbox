data "netbox_webhook" "example" {
  name = "example"
}

output "webhook_id" {
  value = data.netbox_webhook.example.id
}
