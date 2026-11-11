package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxOwnerDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_owner_group" "test" {
  name = "%[1]s"
}

resource "netbox_group" "test" {
  name = "%[1]s"
}
`, testName)
}

func TestAccNetboxOwner_basic(t *testing.T) {
	testName := testAccGetTestName("owner")
	deps := testAccNetboxOwnerDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.5.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_owner", "users/owners/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_owner" "test" {
  name = "%[1]s"
  description = "test"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_owner.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_owner.test", "description", "test"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_owner" "test" {
  name = "%[1]s"
  group_id = netbox_owner_group.test.id
  user_group_ids = [netbox_group.test.id]
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("netbox_owner.test", "group_id", "netbox_owner_group.test", "id"),
					resource.TestCheckResourceAttr("netbox_owner.test", "user_group_ids.#", "1"),
				),
			},
			{
				ResourceName:      "netbox_owner.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxOwnerDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("owner_ds")
	deps := testAccNetboxOwnerDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.5.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_owner" "test" {
  name = "%[1]s"
  description = "test"
}

data "netbox_owner" "test" {
  depends_on = [netbox_owner.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_owner.test", "id", "netbox_owner.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_owner.test", "name", "netbox_owner.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_owner", &resource.Sweeper{
		Name:         "netbox_owner",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "users/owners/", "name")
		},
	})
}
