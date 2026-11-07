package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxRackRole_basic(t *testing.T) {
	testSlug := "rack_role_basic"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_rack_role" "test" {
  name = "%s"
  slug = "%s"
  color_hex = "111111"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rack_role.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_rack_role.test", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_rack_role.test", "color_hex", "111111"),
				),
			},
			{
				ResourceName:      "netbox_rack_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxRackRole_defaultSlug(t *testing.T) {
	testSlug := "rack_role_defSlug"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_rack_role" "test" {
  name = "%s"
  color_hex = "111111"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rack_role.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_rack_role.test", "slug", getSlug(testName)),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_rack_role", &resource.Sweeper{
		Name:         "netbox_rack_role",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/rack-roles/", "name")
		},
	})
}
