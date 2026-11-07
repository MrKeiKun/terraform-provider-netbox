package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccNetboxVirtualChassis_basic(t *testing.T) {
	testSlug := "virtual_chassis"
	testName := testAccGetTestName(testSlug)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualChassisDestroy,
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
