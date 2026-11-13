package netbox

import (
	"fmt"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxRegion_basic(t *testing.T) {
	testSlug := "region_basic"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_region" "test" {
  name = "%s"
  slug = "%s"
  description = "%[1]s"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_region.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_region.test", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_region.test", "parent_region_id", "0"),
					resource.TestCheckResourceAttr("netbox_region.test", "description", testName),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_region" "test" {
  name = "%s"
  slug = "%s"
  description = "%[1]s-updated"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_region.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_region.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxRegion_defaultSlug(t *testing.T) {
	testSlug := "region_defSlug"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_region" "test" {
  name = "%s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_region.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_region.test", "slug", fwresource.Slug(testName)),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_region", &resource.Sweeper{
		Name:         "netbox_region",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/regions/", "name")
		},
	})
}
