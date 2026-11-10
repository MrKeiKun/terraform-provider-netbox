package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxDevicePowerOutletFullDependencies(testName string) string {
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

resource "netbox_device_power_port" "test" {
  device_id = netbox_device.test.id
  name = "%[1]s"
}
`, testName)
}

func TestAccNetboxDevicePowerOutlet_basic(t *testing.T) {
	testSlug := "device_power_outlet_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckDevicePowerOutletDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxDevicePowerOutletFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_device_power_outlet" "test" {
  device_id = netbox_device.test.id
  name = "%[1]s"

	module_id = netbox_module.test.id
	label = "%[1]s_label"
	type = "iec-60320-c5"
	power_port_id = netbox_device_power_port.test.id
	feed_leg = "A"
	description = "%[1]s_description"
	mark_connected = true
  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "label", testName+"_label"),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "type", "iec-60320-c5"),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "feed_leg", "A"),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "description", testName+"_description"),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "mark_connected", "true"),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "tags.0", testName+"a"),

					resource.TestCheckResourceAttrPair("netbox_device_power_outlet.test", "device_id", "netbox_device.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_device_power_outlet.test", "module_id", "netbox_module.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_device_power_outlet.test", "power_port_id", "netbox_device_power_port.test", "id"),
				),
			},
			{
				Config: testAccNetboxDevicePowerOutletFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_device_power_outlet" "test" {
  device_id = netbox_device.test.id
  name = "%[1]s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "label", ""),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "type", ""),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "feed_leg", ""),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "description", ""),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "mark_connected", "false"),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "tags.#", "0"),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "module_id", "0"),
					resource.TestCheckResourceAttr("netbox_device_power_outlet.test", "power_port_id", "0"),

					resource.TestCheckResourceAttrPair("netbox_device_power_outlet.test", "device_id", "netbox_device.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_device_power_outlet.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetboxDevicePowerOutlet_cabledPowerPort is a regression test for the
// NestedPowerPort.cable unmarshal bug: reading a power outlet whose parent
// power port is cabled used to fail with
//
//	json: cannot unmarshal object into Go struct field
//	NestedPowerPort.power_port.cable of type int64
//
// because NetBox 4.x returns the nested power_port's `cable` as an object. The
// second step re-reads (refreshes) the outlet while the cable exists, which is
// exactly the read path that previously errored.
func TestAccNetboxDevicePowerOutlet_cabledPowerPort(t *testing.T) {
	testSlug := "power_outlet_cabled_port"
	testName := testAccGetTestName(testSlug)

	cabledConfig := testAccNetboxDevicePowerOutletFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_power_panel" "test" {
  name    = "%[1]s"
  site_id = netbox_site.test.id
}

resource "netbox_power_feed" "test" {
  power_panel_id          = netbox_power_panel.test.id
  name                    = "%[1]s"
  status                  = "active"
  type                    = "primary"
  supply                  = "ac"
  phase                   = "single-phase"
  voltage                 = 230
  amperage                = 16
  max_percent_utilization = 80
}

resource "netbox_device_power_outlet" "test" {
  device_id     = netbox_device.test.id
  name          = "%[1]s"
  power_port_id = netbox_device_power_port.test.id
}

resource "netbox_cable" "test" {
  a_termination {
    object_type = "dcim.powerfeed"
    object_id   = netbox_power_feed.test.id
  }
  b_termination {
    object_type = "dcim.powerport"
    object_id   = netbox_device_power_port.test.id
  }
  status = "connected"
}`, testName)

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckDevicePowerOutletDestroy,
		Steps: []resource.TestStep{
			{
				Config: cabledConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("netbox_device_power_outlet.test", "power_port_id", "netbox_device_power_port.test", "id"),
				),
			},
			{
				// Re-reading the outlet while its power port is cabled is what
				// used to trip the unmarshal error.
				Config: cabledConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("netbox_device_power_outlet.test", "power_port_id", "netbox_device_power_port.test", "id"),
				),
			},
		},
	})
}

func testAccCheckDevicePowerOutletDestroy(s *terraform.State) error {
	return testAccCheckObjectsDestroyed(s, "netbox_device_power_outlet", "dcim/power-outlets/")
}

func init() {
	resource.AddTestSweepers("netbox_device_power_outlet", &resource.Sweeper{
		Name:         "netbox_device_power_outlet",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/power-outlets/", "name")
		},
	})
}
