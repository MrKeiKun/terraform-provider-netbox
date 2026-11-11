package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxCoolingIntakeDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = "%[1]s"
  status = "active"
}

resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_device_type" "test" {
  model           = "%[1]s"
  manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_device_role" "test" {
  name      = "%[1]s"
  color_hex = "123456"
}

resource "netbox_device" "test" {
  name           = "%[1]s"
  device_type_id = netbox_device_type.test.id
  role_id        = netbox_device_role.test.id
  site_id        = netbox_site.test.id
}
`, testName)
}

func TestAccNetboxCoolingIntake_basic(t *testing.T) {
	testName := testAccGetTestName("coolingintake")
	deps := testAccNetboxCoolingIntakeDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_cooling_intake", "dcim/cooling-intakes/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_intake" "test" {
  name = "%[1]s"
  device_id = netbox_device.test.id
  type = "qdc"
  diameter = 10
  diameter_unit = "mm"
  max_flow = 5
  max_flow_unit = "gpm"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_intake.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cooling_intake.test", "type", "qdc"),
					resource.TestCheckResourceAttr("netbox_cooling_intake.test", "diameter", "10"),
					resource.TestCheckResourceAttr("netbox_cooling_intake.test", "diameter_unit", "mm"),
					resource.TestCheckResourceAttr("netbox_cooling_intake.test", "max_flow", "5"),
					resource.TestCheckResourceAttr("netbox_cooling_intake.test", "max_flow_unit", "gpm"),
					resource.TestCheckResourceAttrPair("netbox_cooling_intake.test", "device_id", "netbox_device.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_intake" "test" {
  description = "updated"
  name = "%[1]s"
  device_id = netbox_device.test.id
  type = "qdc"
  diameter = 10
  diameter_unit = "mm"
  max_flow = 5
  max_flow_unit = "gpm"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_intake.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_cooling_intake.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCoolingIntakeDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("coolingintake_ds")
	deps := testAccNetboxCoolingIntakeDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_intake" "test" {
  name = "%[1]s"
  device_id = netbox_device.test.id
  type = "qdc"
  diameter = 10
  diameter_unit = "mm"
  max_flow = 5
  max_flow_unit = "gpm"
}

data "netbox_cooling_intake" "test" {
  depends_on = [netbox_cooling_intake.test]
  name = "%[1]s"
  device_id = netbox_device.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_cooling_intake.test", "id", "netbox_cooling_intake.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_cooling_intake.test", "name", "netbox_cooling_intake.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_cooling_intake", &resource.Sweeper{
		Name:         "netbox_cooling_intake",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/cooling-intakes/", "name")
		},
	})
}
