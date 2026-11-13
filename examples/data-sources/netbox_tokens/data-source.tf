data "netbox_tokens" "example" {
  filter {
    name  = "created"
    value = "example"
  }
  limit = 10
}

output "tokens_ids" {
  value = data.netbox_tokens.example.tokens[*].id
}
