data "netbox_users" "example" {
  filter {
    name  = "date_joined"
    value = "example"
  }
  limit = 10
}

output "users_ids" {
  value = data.netbox_users.example.users[*].id
}
