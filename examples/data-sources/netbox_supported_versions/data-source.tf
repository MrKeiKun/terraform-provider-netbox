data "netbox_supported_versions" "example" {}

output "supported_versions_id" {
  value = data.netbox_supported_versions.example.id
}
