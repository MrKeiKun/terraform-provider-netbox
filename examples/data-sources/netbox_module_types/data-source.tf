data "netbox_module_types" "example" {
  filter {
    name  = "airflow"
    value = "example"
  }
  limit = 10
}

output "module_types_ids" {
  value = data.netbox_module_types.example.module_types[*].id
}
