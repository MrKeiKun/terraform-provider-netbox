data "netbox_notifications" "example" {
  filter {
    name  = "start"
    value = "example"
  }
  limit = 10
}

output "notifications_ids" {
  value = data.netbox_notifications.example.notifications[*].id
}
