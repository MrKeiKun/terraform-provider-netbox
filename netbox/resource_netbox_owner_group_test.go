package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxOwnerGroupDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxOwnerGroup_basic(t *testing.T) {
	testName := testAccGetTestName("ownergroup")
	deps := testAccNetboxOwnerGroupDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.5.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_owner_group", ownerGroupAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_owner_group" "test" {
  name = "%[1]s"
  description = "test"

}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_owner_group.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_owner_group.test", "description", "test"),
				),
			},
			{
				ResourceName:      "netbox_owner_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxOwnerGroupDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("ownergroup_ds")
	deps := testAccNetboxOwnerGroupDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.5.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_owner_group" "test" {
  name = "%[1]s"
  description = "test"

}

data "netbox_owner_group" "test" {
  depends_on = [netbox_owner_group.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_owner_group.test", "id", "netbox_owner_group.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_owner_group.test", "name", "netbox_owner_group.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_owner_group", &resource.Sweeper{
		Name:         "netbox_owner_group",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "users/owner-groups/", "name")
		},
	})
}
