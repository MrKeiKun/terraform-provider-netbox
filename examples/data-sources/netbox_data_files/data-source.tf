data "netbox_data_files" "example" {
  filter {
    name  = "created"
    value = "example"
  }
  limit = 10
}

output "data_files_ids" {
  value = data.netbox_data_files.example.data_files[*].id
}
