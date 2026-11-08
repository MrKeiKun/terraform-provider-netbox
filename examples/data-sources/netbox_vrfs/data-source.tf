data "netbox_vrfs" "example" {
  filter {
    name  = "name"
    value = "cust-a-prod"
  }
}

output "vrfs" {
  value = data.netbox_vrfs.example.vrfs
}
