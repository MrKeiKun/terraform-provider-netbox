data "netbox_table_config" "example" {
  name = "example"
}

output "table_config_id" {
  value = data.netbox_table_config.example.id
}
