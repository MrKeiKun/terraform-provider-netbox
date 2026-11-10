package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxDevicePowerPortFullDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name = "%[1]s"
}

resource "netbox_site" "test" {
  name = "%[1]s"
  status = "active"
}

resource "netbox_tag" "test" {
  name = "%[1]sa"
}

resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_device_type" "test" {
  model = "%[1]s"
  manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_device_role" "test" {
  name = "%[1]s"
  color_hex = "123456"
}

resource "netbox_device" "test" {
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  tenant_id = netbox_tenant.test.id
  role_id = netbox_device_role.test.id
  site_id = netbox_site.test.id
}

resource "netbox_device_module_bay" "test" {
  device_id = netbox_device.test.id
  name = "%[1]s"
}

resource "netbox_module_type" "test" {
  manufacturer_id = netbox_manufacturer.test.id
  model = "%[1]s"
}

resource "netbox_module" "test" {
  device_id = netbox_device.test.id
  module_bay_id = netbox_device_module_bay.test.id
  module_type_id = netbox_module_type.test.id
  status = "active"
}
`, testName)
}

func TestAccNetboxDevicePowerPort_basic(t *testing.T) {
	testSlug := "device_power_port_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckDevicePowerPortDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxDevicePowerPortFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_device_power_port" "test" {
  device_id = netbox_device.test.id
  name = "%[1]s"

	maximum_draw = 750
  allocated_draw = 500
	mark_connected = true
	type = "iec-60320-c6"
  module_id = netbox_module.test.id
  description = "%[1]s_description"
  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "maximum_draw", "750"),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "allocated_draw", "500"),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "type", "iec-60320-c6"),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "mark_connected", "true"),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "description", testName+"_description"),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "tags.0", testName+"a"),

					resource.TestCheckResourceAttrPair("netbox_device_power_port.test", "device_id", "netbox_device.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_device_power_port.test", "module_id", "netbox_module.test", "id"),
				),
			},
			{
				Config: testAccNetboxDevicePowerPortFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_device_power_port" "test" {
  device_id = netbox_device.test.id
  name = "%[1]s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "maximum_draw", "0"),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "allocated_draw", "0"),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "type", ""),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "mark_connected", "false"),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "description", ""),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "tags.#", "0"),
					resource.TestCheckResourceAttr("netbox_device_power_port.test", "module_id", "0"),

					resource.TestCheckResourceAttrPair("netbox_device_power_port.test", "device_id", "netbox_device.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_device_power_port.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDevicePowerPortDestroy(s *terraform.State) error {
	return testAccCheckObjectsDestroyed(s, "netbox_device_power_port", "dcim/power-ports/")
}

func init() {
	resource.AddTestSweepers("netbox_device_power_port", &resource.Sweeper{
		Name:         "netbox_device_power_port",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/power-ports/", "name")
		},
	})
}
