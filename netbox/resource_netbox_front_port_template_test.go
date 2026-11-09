package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxFrontPortTemplateDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_device_type" "test" {
  model           = "%[1]s"
  manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_rear_port_template" "test" {
  name           = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type           = "8p8c"
  positions      = 1
}
`, testName)
}

func TestAccNetboxFrontPortTemplate_basic(t *testing.T) {
	testName := testAccGetTestName("frontporttempl")
	deps := testAccNetboxFrontPortTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_front_port_template", frontPortTemplateAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_front_port_template" "test" {
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type = "8p8c"
  rear_port_id = netbox_rear_port_template.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_front_port_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_front_port_template.test", "type", "8p8c"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_front_port_template" "test" {
  description = "updated"
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type = "8p8c"
  rear_port_id = netbox_rear_port_template.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_front_port_template.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_front_port_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_front_port_template", &resource.Sweeper{
		Name:         "netbox_front_port_template",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/front-port-templates/", "name")
		},
	})
}
