data "netbox_power_panel" "example" {
  name    = "example"
  site_id = 1
}

output "power_panel_id" {
  value = data.netbox_power_panel.example.id
}
