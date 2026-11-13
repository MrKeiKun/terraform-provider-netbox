data "netbox_rirs" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "rirs_ids" {
  value = data.netbox_rirs.example.rirs[*].id
}
