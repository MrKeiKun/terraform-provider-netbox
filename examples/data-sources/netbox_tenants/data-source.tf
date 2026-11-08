data "netbox_tenants" "example" {
  filter {
    name  = "name"
    value = "customer-a"
  }
}

output "tenants" {
  value = data.netbox_tenants.example.tenants
}
