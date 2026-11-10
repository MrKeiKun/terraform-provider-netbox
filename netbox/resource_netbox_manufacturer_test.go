package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxManufacturer_basic(t *testing.T) {
	testSlug := "manufacturer"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name = "%s"
  slug = "%s"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "slug", randomSlug),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  description = "updated"
  name = "%s"
  slug = "%s"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_manufacturer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_manufacturer", &resource.Sweeper{
		Name:         "netbox_manufacturer",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/manufacturers/", "name")
		},
	})
}
