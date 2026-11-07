package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxGroup_basic(t *testing.T) {
	testSlug := "groups"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_group" "test_basic" {
  name = "%s"
	description = "This is my example resource"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_group.test_basic", "name", testName),
					resource.TestCheckResourceAttr("netbox_group.test_basic", "description", "This is my example resource"),
				),
			},
			{
				ResourceName:      "netbox_group.test_basic",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_group", &resource.Sweeper{
		Name:         "netbox_group",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "users/groups/", "name")
		},
	})
}
