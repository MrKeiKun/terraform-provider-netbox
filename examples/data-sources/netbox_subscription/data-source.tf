data "netbox_subscription" "example" {
  id = 1
}

output "subscription_id" {
  value = data.netbox_subscription.example.id
}
