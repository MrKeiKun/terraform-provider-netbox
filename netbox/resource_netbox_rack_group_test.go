package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxRackGroupDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxRackGroup_basic(t *testing.T) {
	testName := testAccGetTestName("rackgroup")
	deps := testAccNetboxRackGroupDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.6.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_rack_group", rackGroupAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_rack_group" "test" {
  name = "%[1]s"
  description = "test"

}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rack_group.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_rack_group.test", "description", "test"),
					resource.TestCheckResourceAttrSet("netbox_rack_group.test", "slug"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_rack_group" "test" {
  name = "%[1]s"
  description = "test-updated"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_rack_group.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_rack_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxRackGroupDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("rackgroup_ds")
	deps := testAccNetboxRackGroupDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.6.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_rack_group" "test" {
  name = "%[1]s"
  description = "test"

}

data "netbox_rack_group" "test" {
  depends_on = [netbox_rack_group.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_rack_group.test", "id", "netbox_rack_group.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_rack_group.test", "name", "netbox_rack_group.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_rack_group", &resource.Sweeper{
		Name:         "netbox_rack_group",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/rack-groups/", "name")
		},
	})
}
