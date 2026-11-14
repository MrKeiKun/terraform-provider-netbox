package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxVirtualChassisMaster_basic(t *testing.T) {
	testName := testAccGetTestName("virtual_chassis_master")
	config := fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = "%[1]s"
  status = "active"
}

resource "netbox_device_role" "test" {
  name      = "%[1]s"
  color_hex = "123456"
}

resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_device_type" "test" {
  model           = "%[1]s"
  manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_virtual_chassis" "test" {
  name = "%[1]s"
}

resource "netbox_device" "member" {
  name                     = "%[1]s"
  role_id                  = netbox_device_role.test.id
  site_id                  = netbox_site.test.id
  device_type_id           = netbox_device_type.test.id
  virtual_chassis_id       = netbox_virtual_chassis.test.id
  virtual_chassis_position = 1
  virtual_chassis_master   = false

  lifecycle {
    ignore_changes = [virtual_chassis_master]
  }
}

resource "netbox_virtual_chassis_master" "test" {
  virtual_chassis_id = netbox_virtual_chassis.test.id
  master_id          = netbox_device.member.id
}
`, testName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("netbox_virtual_chassis_master.test", "virtual_chassis_id", "netbox_virtual_chassis.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_virtual_chassis_master.test", "master_id", "netbox_device.member", "id"),
				),
			},
			{
				ResourceName:      "netbox_virtual_chassis_master.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
