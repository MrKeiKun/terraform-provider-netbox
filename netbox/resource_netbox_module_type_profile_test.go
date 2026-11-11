package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxModuleTypeProfileDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxModuleTypeProfile_basic(t *testing.T) {
	testName := testAccGetTestName("moduletypeprof")
	deps := testAccNetboxModuleTypeProfileDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_module_type_profile", "dcim/module-type-profiles/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_module_type_profile" "test" {
  name = "%[1]s"
  description = "test"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_module_type_profile.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_module_type_profile.test", "description", "test"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_module_type_profile" "test" {
  name = "%[1]s"
  description = "test-updated"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_module_type_profile.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_module_type_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxModuleTypeProfileDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("moduletypeprof_ds")
	deps := testAccNetboxModuleTypeProfileDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_module_type_profile" "test" {
  name = "%[1]s"
  description = "test"
}

data "netbox_module_type_profile" "test" {
  depends_on = [netbox_module_type_profile.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_module_type_profile.test", "id", "netbox_module_type_profile.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_module_type_profile.test", "name", "netbox_module_type_profile.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_module_type_profile", &resource.Sweeper{
		Name:         "netbox_module_type_profile",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/module-type-profiles/", "name")
		},
	})
}
