data "netbox_asn_ranges" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "asn_ranges_ids" {
  value = data.netbox_asn_ranges.example.asn_ranges[*].id
}
