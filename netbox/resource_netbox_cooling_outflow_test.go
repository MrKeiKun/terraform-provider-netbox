package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxCoolingOutflowDependencies(testName string) string {
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

func TestAccNetboxCoolingOutflow_basic(t *testing.T) {
	testName := testAccGetTestName("coolingoutflow")
	deps := testAccNetboxCoolingOutflowDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_cooling_outflow", coolingOutflowAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_outflow" "test" {
  name = "%[1]s"
  device_id = netbox_device.test.id
  type = "qdc"
  diameter = 10
  diameter_unit = "mm"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_outflow.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cooling_outflow.test", "type", "qdc"),
					resource.TestCheckResourceAttr("netbox_cooling_outflow.test", "diameter", "10"),
					resource.TestCheckResourceAttr("netbox_cooling_outflow.test", "diameter_unit", "mm"),
					resource.TestCheckResourceAttrPair("netbox_cooling_outflow.test", "device_id", "netbox_device.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_cooling_outflow.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCoolingOutflowDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("coolingoutflow_ds")
	deps := testAccNetboxCoolingOutflowDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_outflow" "test" {
  name = "%[1]s"
  device_id = netbox_device.test.id
  type = "qdc"
  diameter = 10
  diameter_unit = "mm"
}

data "netbox_cooling_outflow" "test" {
  depends_on = [netbox_cooling_outflow.test]
  name = "%[1]s"
  device_id = netbox_device.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_cooling_outflow.test", "id", "netbox_cooling_outflow.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_cooling_outflow.test", "name", "netbox_cooling_outflow.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_cooling_outflow", &resource.Sweeper{
		Name:         "netbox_cooling_outflow",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/cooling-outflows/", "name")
		},
	})
}
