resource "netbox_circuit_provider" "test" {
  name = "test"
}

resource "netbox_provider_network" "test" {
  name        = "test_network"
  provider_id = netbox_circuit_provider.test.id
  service_id  = "SVC001"
  description = "Test provider network"
}