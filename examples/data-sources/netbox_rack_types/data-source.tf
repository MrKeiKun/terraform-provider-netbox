data "netbox_rack_types" "example" {
  filter {
    name  = "slug"
    value = "example"
  }
  limit = 10
}

output "rack_types_ids" {
  value = data.netbox_rack_types.example.rack_types[*].id
}
