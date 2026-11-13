data "netbox_journal_entries" "example" {
  filter {
    name  = "assigned_object_id"
    value = "example"
  }
  limit = 10
}

output "journal_entries_ids" {
  value = data.netbox_journal_entries.example.journal_entries[*].id
}
