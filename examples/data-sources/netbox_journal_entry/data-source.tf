data "netbox_journal_entry" "example" {
  id = 1
}

output "journal_entry_id" {
  value = data.netbox_journal_entry.example.id
}
