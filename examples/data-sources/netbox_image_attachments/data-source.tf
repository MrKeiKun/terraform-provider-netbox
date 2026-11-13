data "netbox_image_attachments" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "image_attachments_ids" {
  value = data.netbox_image_attachments.example.image_attachments[*].id
}
