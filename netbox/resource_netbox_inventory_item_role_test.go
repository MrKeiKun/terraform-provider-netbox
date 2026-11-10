package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxInventoryItemRoleFullDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%[1]sa"
}
`, testName)
}

func TestAccNetboxInventoryItemRole_basic(t *testing.T) {
	testSlug := "inventory_item_role_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckInventoryItemRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxInventoryItemRoleFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_inventory_item_role" "test" {
  name = "%[1]s"
  slug = "%[1]s_slug"
	color_hex = "123456"

  description = "%[1]s_description"
  tags = [ netbox_tag.test.name ]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "slug", testName+"_slug"),
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "description", testName+"_description"),
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "color_hex", "123456"),
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "tags.0", testName+"a"),
				),
			},
			{
				Config: testAccNetboxInventoryItemRoleFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_inventory_item_role" "test" {
  name = "%[1]s"
  slug = "%[1]s_slug"
	color_hex = "123456"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "slug", testName+"_slug"),
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "description", ""),
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "color_hex", "123456"),
					resource.TestCheckResourceAttr("netbox_inventory_item_role.test", "tags.#", "0"),
				),
			},
			{
				ResourceName:      "netbox_inventory_item_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckInventoryItemRoleDestroy(s *terraform.State) error {
	return testAccCheckObjectsDestroyed(s, "netbox_inventory_item_role", "dcim/inventory-item-roles/")
}

func init() {
	resource.AddTestSweepers("netbox_inventory_item_role", &resource.Sweeper{
		Name:         "netbox_inventory_item_role",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/inventory-item-roles/", "name")
		},
	})
}
