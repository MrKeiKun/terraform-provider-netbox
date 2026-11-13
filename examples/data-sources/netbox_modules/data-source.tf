data "netbox_modules" "example" {
  filter {
    name  = "status"
    value = "example"
  }
  limit = 10
}

output "modules_ids" {
  value = data.netbox_modules.example.modules[*].id
}
