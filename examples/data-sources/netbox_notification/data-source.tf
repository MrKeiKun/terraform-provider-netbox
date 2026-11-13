data "netbox_notification" "example" {
  id = 1
}

output "notification_id" {
  value = data.netbox_notification.example.id
}
