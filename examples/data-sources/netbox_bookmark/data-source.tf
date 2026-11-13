data "netbox_bookmark" "example" {
  id = 1
}

output "bookmark_id" {
  value = data.netbox_bookmark.example.id
}
