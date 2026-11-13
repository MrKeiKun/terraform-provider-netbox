data "netbox_platforms" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "platforms_ids" {
  value = data.netbox_platforms.example.platforms[*].id
}
