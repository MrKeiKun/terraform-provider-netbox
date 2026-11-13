data "netbox_table_configs" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "table_configs_ids" {
  value = data.netbox_table_configs.example.table_configs[*].id
}
