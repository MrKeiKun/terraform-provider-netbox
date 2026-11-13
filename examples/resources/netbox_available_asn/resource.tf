resource "netbox_rir" "example" {
  name = "Example RIR"
  slug = "example-rir"
}

resource "netbox_asn_range" "example" {
  name   = "Example range"
  slug   = "example-range"
  rir_id = netbox_rir.example.id
  start  = 64512
  end    = 65534
}

// NetBox picks the next free AS number of the range
resource "netbox_available_asn" "example" {
  asn_range_id = netbox_asn_range.example.id
  description  = "Allocated by Terraform"
}
