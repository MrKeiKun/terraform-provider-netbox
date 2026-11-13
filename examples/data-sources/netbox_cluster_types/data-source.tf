data "netbox_cluster_types" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "cluster_types_ids" {
  value = data.netbox_cluster_types.example.cluster_types[*].id
}
