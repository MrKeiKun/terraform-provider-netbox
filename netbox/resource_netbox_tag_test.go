package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxTag_basic(t *testing.T) {
	testSlug := "tag_basic"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%s"
  slug = "%s"
  color_hex = "112233"
  description = "This is a test"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tag.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_tag.test", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_tag.test", "color_hex", "112233"),
					resource.TestCheckResourceAttr("netbox_tag.test", "description", "This is a test"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%s"
  slug = "%s"
  color_hex = "112233"
  description = "This is a test-updated"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_tag.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxTag_defaultSlug(t *testing.T) {
	testSlug := "tag_defSlug"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tag.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_tag.test", "slug", getSlug(testName)),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_tag", &resource.Sweeper{
		Name:         "netbox_tag",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "extras/tags/", "name")
		},
	})
}
