package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxInventoryItemTemplateDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_device_type" "test" {
  model           = "%[1]s"
  manufacturer_id = netbox_manufacturer.test.id
}
`, testName)
}

func TestAccNetboxInventoryItemTemplate_basic(t *testing.T) {
	testName := testAccGetTestName("inventoryitemt")
	deps := testAccNetboxInventoryItemTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_inventory_item_template", inventoryItemTemplateAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_inventory_item_template" "test" {
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  part_id = "abc"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_inventory_item_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_inventory_item_template.test", "part_id", "abc"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_inventory_item_template" "test" {
  description = "updated"
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  part_id = "abc"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_inventory_item_template.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_inventory_item_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxInventoryItemTemplateDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("inventoryitemt_ds")
	deps := testAccNetboxInventoryItemTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_inventory_item_template" "test" {
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  part_id = "abc"
}

data "netbox_inventory_item_template" "test" {
  depends_on = [netbox_inventory_item_template.test]
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_inventory_item_template.test", "id", "netbox_inventory_item_template.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_inventory_item_template.test", "name", "netbox_inventory_item_template.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_inventory_item_template", &resource.Sweeper{
		Name:         "netbox_inventory_item_template",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/inventory-item-templates/", "name")
		},
	})
}
