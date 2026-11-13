data "netbox_image_attachment" "example" {
  name = "example"
}

output "image_attachment_id" {
  value = data.netbox_image_attachment.example.id
}
