data "netbox_prefixes" "example" {
  filter {
    name  = "vrf_id"
    value = "1"
  }
}

output "prefixes" {
  value = data.netbox_prefixes.example.prefixes
}
