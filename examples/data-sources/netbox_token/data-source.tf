data "netbox_token" "example" {
  id = 1
}

output "token_id" {
  value = data.netbox_token.example.id
}
