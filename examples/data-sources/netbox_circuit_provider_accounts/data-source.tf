data "netbox_circuit_provider_accounts" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "circuit_provider_accounts_ids" {
  value = data.netbox_circuit_provider_accounts.example.circuit_provider_accounts[*].id
}
