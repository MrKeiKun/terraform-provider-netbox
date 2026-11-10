package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxRearPortTemplate_basic(t *testing.T) {
	testSlug := "rear_port_template"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
	name = "%[1]s"
}

resource "netbox_device_type" "test" {
	model = "%[1]s"
	slug = "%[2]s"
	part_number = "%[2]s"
	manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_rear_port_template" "test" {
	name = "%[1]s"
	device_type_id = netbox_device_type.test.id
	type = "8p8c"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "type", "8p8c"),
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "positions", "1"),
					resource.TestCheckResourceAttrPair("netbox_rear_port_template.test", "device_type_id", "netbox_device_type.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
	name = "%[1]s"
}

resource "netbox_device_type" "test" {
	model = "%[1]s"
	slug = "%[2]s"
	part_number = "%[2]s"
	manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_rear_port_template" "test" {
  description = "updated"
	name = "%[1]s"
	device_type_id = netbox_device_type.test.id
	type = "8p8c"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_rear_port_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxRearPortTemplate_opts(t *testing.T) {
	testSlug := "rear_port_template"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
	name = "%[1]s"
}

resource "netbox_device_type" "test" {
	model = "%[1]s"
	slug = "%[2]s"
	part_number = "%[2]s"
	manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_rear_port_template" "test" {
	name = "%[1]s"
	description = "%[1]s description"
	label = "%[1]s label"
	device_type_id = netbox_device_type.test.id
	type = "mpo"
	positions = 12
	color_hex = "f44336"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "description", fmt.Sprintf("%s description", testName)),
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "label", fmt.Sprintf("%s label", testName)),
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "type", "mpo"),
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "positions", "12"),
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "color_hex", "f44336"),
					resource.TestCheckResourceAttrPair("netbox_rear_port_template.test", "device_type_id", "netbox_device_type.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
	name = "%[1]s"
}

resource "netbox_module_type" "test" {
	manufacturer_id = netbox_manufacturer.test.id
	model           = "%[1]s"
}

resource "netbox_rear_port_template" "test" {
	name = "%[1]s"
	description = "%[1]s description"
	label = "%[1]s label"
	module_type_id = netbox_module_type.test.id
	type = "lc-upc"
	positions = 6
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "type", "lc-upc"),
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "positions", "6"),
					resource.TestCheckResourceAttrPair("netbox_rear_port_template.test", "module_type_id", "netbox_module_type.test", "id"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_rear_port_template", &resource.Sweeper{
		Name:         "netbox_rear_port_template",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/rear-port-templates/", "name")
		},
	})
}

// A rear port template that has a front port mapped to it must still be
// updatable: NetBox rejects an update that sends the existing mapping again.
func TestAccNetboxRearPortTemplate_updateWithFrontPort(t *testing.T) {
	testName := testAccGetTestName("rpt_mapped")
	config := func(description string) string {
		return fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_device_type" "test" {
  model = "%[1]s"
  manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_rear_port_template" "test" {
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type = "8p8c"
  description = "%[2]s"
}

resource "netbox_front_port_template" "test" {
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type = "8p8c"
  rear_port_id = netbox_rear_port_template.test.id
}`, testName, description)
	}
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config("before"),
				Check:  resource.TestCheckResourceAttr("netbox_rear_port_template.test", "description", "before"),
			},
			{
				Config: config("after"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rear_port_template.test", "description", "after"),
					testAccCheckIfNetboxAtLeast("4.5.0", resource.TestCheckResourceAttr("netbox_rear_port_template.test", "front_ports.#", "1")),
				),
			},
		},
	})
}
