resource "netbox_provider_account" "example" {
  provider_id = netbox_circuit_provider.example.id
  account     = "example"
}
