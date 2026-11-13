data "netbox_permission" "example" {
  name = "example"
}

output "permission_id" {
  value = data.netbox_permission.example.id
}
