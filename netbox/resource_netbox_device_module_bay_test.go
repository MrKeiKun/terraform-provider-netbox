package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxDeviceModuleBayFullDependencies(testName string) string {
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
`, testName)
}

func TestAccNetboxDeviceModuleBay_basic(t *testing.T) {
	testSlug := "device_module_bay_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckDeviceModuleBayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxDeviceModuleBayFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_device_module_bay" "test" {
  device_id = netbox_device.test.id
  name = "%[1]s"
  label = "%[1]s_label"
  position = "testposition"
  description = "%[1]s_description"
  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "label", testName+"_label"),
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "position", "testposition"),
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "description", testName+"_description"),
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "tags.0", testName+"a"),

					resource.TestCheckResourceAttrPair("netbox_device_module_bay.test", "device_id", "netbox_device.test", "id"),
				),
			},
			{
				Config: testAccNetboxDeviceModuleBayFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_device_module_bay" "test" {
  device_id = netbox_device.test.id
  name = "%[1]s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "label", ""),
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "position", ""),
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "description", ""),
					resource.TestCheckResourceAttr("netbox_device_module_bay.test", "tags.#", "0"),

					resource.TestCheckResourceAttrPair("netbox_device_module_bay.test", "device_id", "netbox_device.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_device_module_bay.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDeviceModuleBayDestroy(s *terraform.State) error {
	return testAccCheckObjectsDestroyed(s, "netbox_device_module_bay", "dcim/module-bays/")
}

func init() {
	resource.AddTestSweepers("netbox_device_module_bay", &resource.Sweeper{
		Name:         "netbox_device_module_bay",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/module-bays/", "name")
		},
	})
}
