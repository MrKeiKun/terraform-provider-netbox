data "netbox_l2vpns" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "l2vpns_ids" {
  value = data.netbox_l2vpns.example.l2vpns[*].id
}
