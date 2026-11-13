data "netbox_group" "example" {
  name = "example"
}

output "group_id" {
  value = data.netbox_group.example.id
}
