data "netbox_event_rule" "example" {
  name = "example"
}

output "event_rule_id" {
  value = data.netbox_event_rule.example.id
}
