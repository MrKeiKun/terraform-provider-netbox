data "netbox_circuit_provider_account" "example" {
  account = "example"
}

output "circuit_provider_account_id" {
  value = data.netbox_circuit_provider_account.example.id
}
