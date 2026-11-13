data "netbox_user" "example" {
  id = 1
}

output "user_id" {
  value = data.netbox_user.example.id
}
