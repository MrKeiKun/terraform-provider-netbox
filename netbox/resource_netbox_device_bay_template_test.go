package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxDeviceBayTemplate_basic(t *testing.T) {
	testSlug := "device_bay_template"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
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
	subdevice_role = "parent"
}

resource "netbox_device_bay_template" "test" {
	name = "%[1]s"
	device_type_id = netbox_device_type.test.id
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_bay_template.test", "name", testName),
					resource.TestCheckResourceAttrPair("netbox_device_bay_template.test", "device_type_id", "netbox_device_type.test", "id"),
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
	subdevice_role = "parent"
}

resource "netbox_device_bay_template" "test" {
  description = "updated"
	name = "%[1]s"
	device_type_id = netbox_device_type.test.id
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_bay_template.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_device_bay_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxDeviceBayTemplate_opts(t *testing.T) {
	testSlug := "device_bay_template"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
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
	subdevice_role = "parent"
	manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_device_bay_template" "test" {
	name = "%[1]s"
	description = "%[1]s description"
	label = "%[1]s label"
	device_type_id = netbox_device_type.test.id
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_bay_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_device_bay_template.test", "description", fmt.Sprintf("%s description", testName)),
					resource.TestCheckResourceAttr("netbox_device_bay_template.test", "label", fmt.Sprintf("%s label", testName)),
					resource.TestCheckResourceAttrPair("netbox_device_bay_template.test", "device_type_id", "netbox_device_type.test", "id"),
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
	subdevice_role = "parent"
	manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_device_bay_template" "test" {
	name = "%[1]s"
	description = "%[1]s description"
	label = "%[1]s label"
	device_type_id = netbox_device_type.test.id
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_bay_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_device_bay_template.test", "description", fmt.Sprintf("%s description", testName)),
					resource.TestCheckResourceAttr("netbox_device_bay_template.test", "label", fmt.Sprintf("%s label", testName)),
					resource.TestCheckResourceAttrPair("netbox_device_bay_template.test", "device_type_id", "netbox_device_type.test", "id"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_device_bay_template", &resource.Sweeper{
		Name:         "netbox_device_bay_template",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/device-bay-templates/", "name")
		},
	})
}
