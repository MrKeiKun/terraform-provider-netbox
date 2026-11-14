resource "netbox_virtual_chassis" "example" {
  name = "example"
}

resource "netbox_device" "member" {
  name                     = "member"
  device_type_id           = netbox_device_type.example.id
  role_id                  = netbox_device_role.example.id
  site_id                  = netbox_site.example.id
  virtual_chassis_id       = netbox_virtual_chassis.example.id
  virtual_chassis_position = 1
  virtual_chassis_master   = false

  lifecycle {
    ignore_changes = [virtual_chassis_master]
  }
}

resource "netbox_virtual_chassis_master" "example" {
  virtual_chassis_id = netbox_virtual_chassis.example.id
  master_id          = netbox_device.member.id
}
