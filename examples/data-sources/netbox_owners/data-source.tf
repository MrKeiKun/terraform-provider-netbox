data "netbox_owners" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "owners_ids" {
  value = data.netbox_owners.example.owners[*].id
}
