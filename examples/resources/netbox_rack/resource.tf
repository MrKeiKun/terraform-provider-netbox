resource "netbox_manufacturer" "test" {
  name = "test-manufacturer"
}

resource "netbox_site" "test" {
  name   = "test"
  status = "active"
}

resource "netbox_rack_type" "test" {
  model           = "test-model"
  manufacturer_id = netbox_manufacturer.test.id
  width           = 19
  u_height        = 48
  starting_unit   = 1
  form_factor     = "2-post-frame"
}

resource "netbox_rack" "test" {
  name         = "test"
  site_id      = netbox_site.test.id
  status       = "reserved"
  rack_type_id = netbox_rack_type.test.id
  width        = 19
  u_height     = 48
}
