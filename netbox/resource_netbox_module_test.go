package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxModuleFullDependencies(testName string) string {
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
}`, testName)
}

func TestAccNetboxModule_basic(t *testing.T) {
	testSlug := "module_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxModuleFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_module" "test" {
	device_id = netbox_device.test.id
	module_bay_id = netbox_device_module_bay.test.id
	module_type_id = netbox_module_type.test.id
	status = "active"

	serial = "%[1]s_serial"
	asset_tag = "%[1]s_asset"
	description = "%[1]s_description"
  comments = "%[1]s_comments"
  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_module.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_module.test", "serial", testName+"_serial"),
					resource.TestCheckResourceAttr("netbox_module.test", "asset_tag", testName+"_asset"),
					resource.TestCheckResourceAttr("netbox_module.test", "description", testName+"_description"),
					resource.TestCheckResourceAttr("netbox_module.test", "comments", testName+"_comments"),
					resource.TestCheckResourceAttr("netbox_module.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_module.test", "tags.0", testName+"a"),

					resource.TestCheckResourceAttrPair("netbox_module.test", "device_id", "netbox_device.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_module.test", "module_bay_id", "netbox_device_module_bay.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_module.test", "module_type_id", "netbox_module_type.test", "id"),
				),
			},
			{
				Config: testAccNetboxModuleFullDependencies(testName) + `
resource "netbox_module" "test" {
	device_id = netbox_device.test.id
	module_bay_id = netbox_device_module_bay.test.id
	module_type_id = netbox_module_type.test.id
	status = "offline"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_module.test", "status", "offline"),
					resource.TestCheckResourceAttr("netbox_module.test", "serial", ""),
					resource.TestCheckResourceAttr("netbox_module.test", "asset_tag", ""),
					resource.TestCheckResourceAttr("netbox_module.test", "description", ""),
					resource.TestCheckResourceAttr("netbox_module.test", "comments", ""),
					resource.TestCheckResourceAttr("netbox_module.test", "tags.#", "0"),

					resource.TestCheckResourceAttrPair("netbox_module.test", "device_id", "netbox_device.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_module.test", "module_bay_id", "netbox_device_module_bay.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_module.test", "module_type_id", "netbox_module_type.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_module.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckModuleDestroy(s *terraform.State) error {
	return testAccCheckObjectsDestroyed(s, "netbox_module", "dcim/modules/")
}

func init() {
	resource.AddTestSweepers("netbox_module", &resource.Sweeper{
		Name:         "netbox_module",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/modules/", "id")
		},
	})
}
