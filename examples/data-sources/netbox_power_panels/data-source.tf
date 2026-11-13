data "netbox_power_panels" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "power_panels_ids" {
  value = data.netbox_power_panels.example.power_panels[*].id
}
