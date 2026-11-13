data "netbox_cooling_intake_templates" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "cooling_intake_templates_ids" {
  value = data.netbox_cooling_intake_templates.example.cooling_intake_templates[*].id
}
