data "netbox_cluster_groups" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "cluster_groups_ids" {
  value = data.netbox_cluster_groups.example.cluster_groups[*].id
}
