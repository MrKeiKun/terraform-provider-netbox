data "netbox_tenant_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "tenant_groups_ids" {
  value = data.netbox_tenant_groups.example.tenant_groups[*].id
}
