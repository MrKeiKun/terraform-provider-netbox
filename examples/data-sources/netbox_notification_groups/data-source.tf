data "netbox_notification_groups" "example" {
  filter {
    name  = "start"
    value = "example"
  }
  limit = 10
}

output "notification_groups_ids" {
  value = data.netbox_notification_groups.example.notification_groups[*].id
}
