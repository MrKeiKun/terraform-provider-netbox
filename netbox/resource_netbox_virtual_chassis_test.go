package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccNetboxVirtualChassis_basic(t *testing.T) {
	testSlug := "virtual_chassis"
	testName := testAccGetTestName(testSlug)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVirtualChassisDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "tag_a" {
	name = "[%[1]s_a]"
	color_hex = "123456"
}
resource "netbox_virtual_chassis" "test" {
	name = "%[1]s"
	domain = "domain"
	description = "description"
	comments = "comment"
	tags = [netbox_tag.tag_a.name]
}
				`, testName),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "tag_a" {
	name = "[%[1]s_a]"
	color_hex = "123456"
}
resource "netbox_virtual_chassis" "test" {
	name = "%[1]s_updated"
	domain = "domain_updated"
	description = "description updated"
	comments = "comment updated"
	tags = [netbox_tag.tag_a.name]
}
				`, testName),
			},
			{
				ResourceName:      "netbox_virtual_chassis.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVirtualChassisDestroy(s *terraform.State) error {
	return testAccCheckObjectsDestroyed(s, "netbox_virtual_chassis", "dcim/virtual-chassis/")
}

func TestAccNetboxVirtualChassis_master(t *testing.T) {
	testName := testAccGetTestName("vcmaster")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVirtualChassisDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site" "test" {
  name = "%[1]s"
}

resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_device_role" "test" {
  name      = "%[1]s"
  color_hex = "123456"
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
  site_id                  = netbox_site.test.id
  role_id                  = netbox_device_role.test.id
  device_type_id           = netbox_device_type.test.id
  virtual_chassis_id       = netbox_virtual_chassis.test.id
  virtual_chassis_position = 1
  virtual_chassis_master   = true
}
`, testName),
			},
			{
				RefreshState: true,
				Check:        resource.TestCheckResourceAttrPair("netbox_virtual_chassis.test", "master_id", "netbox_device.member", "id"),
			},
		},
	})
}
