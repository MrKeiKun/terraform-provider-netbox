data "netbox_bookmarks" "example" {
  filter {
    name  = "created"
    value = "example"
  }
  limit = 10
}

output "bookmarks_ids" {
  value = data.netbox_bookmarks.example.bookmarks[*].id
}
