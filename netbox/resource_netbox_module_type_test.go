package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxModuleTypeFullDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_tag" "test" {
	name = "%[1]sa"
}
`, testName)
}

func TestAccNetboxModuleType_basic(t *testing.T) {
	testSlug := "module_type_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckModuleTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxModuleTypeFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_module_type" "test" {
  manufacturer_id = netbox_manufacturer.test.id
  model = "%[1]s"
  part_number = "%[1]s_pn"
  description = "%[1]s_description"
  comments = "%[1]s_comments"

  weight = 1
  weight_unit = "kg"
  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_module_type.test", "model", testName),
					resource.TestCheckResourceAttr("netbox_module_type.test", "part_number", testName+"_pn"),
					resource.TestCheckResourceAttr("netbox_module_type.test", "description", testName+"_description"),
					resource.TestCheckResourceAttr("netbox_module_type.test", "comments", testName+"_comments"),
					resource.TestCheckResourceAttr("netbox_module_type.test", "weight", "1"),
					resource.TestCheckResourceAttr("netbox_module_type.test", "weight_unit", "kg"),
					resource.TestCheckResourceAttr("netbox_module_type.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_module_type.test", "tags.0", testName+"a"),

					resource.TestCheckResourceAttrPair("netbox_module_type.test", "manufacturer_id", "netbox_manufacturer.test", "id"),
				),
			},
			{
				Config: testAccNetboxModuleTypeFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_module_type" "test" {
  manufacturer_id = netbox_manufacturer.test.id
  model = "%[1]s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_module_type.test", "model", testName),
					resource.TestCheckResourceAttr("netbox_module_type.test", "part_number", ""),
					resource.TestCheckResourceAttr("netbox_module_type.test", "description", ""),
					resource.TestCheckResourceAttr("netbox_module_type.test", "comments", ""),
					resource.TestCheckResourceAttr("netbox_module_type.test", "weight", "0"),
					resource.TestCheckResourceAttr("netbox_module_type.test", "weight_unit", ""),
					resource.TestCheckResourceAttr("netbox_module_type.test", "tags.#", "0"),

					resource.TestCheckResourceAttrPair("netbox_module_type.test", "manufacturer_id", "netbox_manufacturer.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_module_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckModuleTypeDestroy(s *terraform.State) error {
	return testAccCheckObjectsDestroyed(s, "netbox_module_type", "dcim/module-types/")
}

func init() {
	resource.AddTestSweepers("netbox_module_type", &resource.Sweeper{
		Name:         "netbox_module_type",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/module-types/", "model")
		},
	})
}
