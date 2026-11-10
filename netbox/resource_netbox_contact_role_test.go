package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxContactRole_basic(t *testing.T) {
	testSlug := "contactrole"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_contact_role" "test" {
  name = "%s"
  slug = "%s"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_contact_role.test", "name", testName),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_contact_role" "test" {
  description = "updated"
  name = "%s"
  slug = "%s"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_contact_role.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_contact_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_contact_role", &resource.Sweeper{
		Name:         "netbox_contact_role",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "tenancy/contact-roles/", "name")
		},
	})
}
