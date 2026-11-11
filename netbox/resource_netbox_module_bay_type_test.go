package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxModuleBayTypeDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxModuleBayType_basic(t *testing.T) {
	testName := testAccGetTestName("modulebaytype")
	deps := testAccNetboxModuleBayTypeDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_module_bay_type", "dcim/module-bay-types/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_module_bay_type" "test" {
  name = "%[1]s"
  color_hex = "ff0000"

}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_module_bay_type.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_module_bay_type.test", "color_hex", "ff0000"),
					resource.TestCheckResourceAttrSet("netbox_module_bay_type.test", "slug"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_module_bay_type" "test" {
  description = "updated"
  name = "%[1]s"
  color_hex = "ff0000"

}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_module_bay_type.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_module_bay_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxModuleBayTypeDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("modulebaytype_ds")
	deps := testAccNetboxModuleBayTypeDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_module_bay_type" "test" {
  name = "%[1]s"
  color_hex = "ff0000"

}

data "netbox_module_bay_type" "test" {
  depends_on = [netbox_module_bay_type.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_module_bay_type.test", "id", "netbox_module_bay_type.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_module_bay_type.test", "name", "netbox_module_bay_type.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_module_bay_type", &resource.Sweeper{
		Name:         "netbox_module_bay_type",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/module-bay-types/", "name")
		},
	})
}
