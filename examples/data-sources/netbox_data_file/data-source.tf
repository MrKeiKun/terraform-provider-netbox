data "netbox_data_file" "example" {
  id = 1
}

output "data_file_id" {
  value = data.netbox_data_file.example.id
}
