data "netbox_tagged_objects" "example" {
  filter {
    name  = "id"
    value = "example"
  }
  limit = 10
}

output "tagged_objects_ids" {
  value = data.netbox_tagged_objects.example.tagged_objects[*].id
}
