data "netbox_event_rules" "example" {
  filter {
    name  = "name"
    value = "example"
  }
  limit = 10
}

output "event_rules_ids" {
  value = data.netbox_event_rules.example.event_rules[*].id
}
