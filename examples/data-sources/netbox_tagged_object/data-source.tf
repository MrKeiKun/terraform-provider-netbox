data "netbox_tagged_object" "example" {
  id = 1
}

output "tagged_object_id" {
  value = data.netbox_tagged_object.example.id
}
