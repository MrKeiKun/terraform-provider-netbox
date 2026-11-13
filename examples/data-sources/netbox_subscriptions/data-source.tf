data "netbox_subscriptions" "example" {
  filter {
    name  = "start"
    value = "example"
  }
  limit = 10
}

output "subscriptions_ids" {
  value = data.netbox_subscriptions.example.subscriptions[*].id
}
