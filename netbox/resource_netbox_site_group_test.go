package netbox

import (
	"fmt"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxSiteGroup_basic(t *testing.T) {
	testSlug := "s_grp_basic"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site_group" "parent" {
  name = "%[1]s"
  slug = "%[2]s"
  description = "foo bar."
}

resource "netbox_site_group" "child" {
  name = "%[1]s-child"
  slug = "%[2]s-c"

  parent_id = netbox_site_group.parent.id
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site_group.parent", "name", testName),
					resource.TestCheckResourceAttr("netbox_site_group.parent", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_site_group.parent", "description", "foo bar."),
					resource.TestCheckResourceAttr("netbox_site_group.child", "name", fmt.Sprintf("%s-child", testName)),
					resource.TestCheckResourceAttr("netbox_site_group.child", "slug", fmt.Sprintf("%s-c", randomSlug)),
					resource.TestCheckResourceAttrPair("netbox_site_group.child", "parent_id", "netbox_site_group.parent", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_site_group" "parent" {
  name = "%[1]s"
  slug = "%[2]s"
  description = "foo bar.-updated"
}

resource "netbox_site_group" "child" {
  name = "%[1]s-child"
  slug = "%[2]s-c"

  parent_id = netbox_site_group.parent.id
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_site_group.parent", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_site_group.parent",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxSiteGroup_defaultSlug(t *testing.T) {
	testSlug := "sitegrp_defSlug"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site_group" "test" {
  name = "%s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site_group.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_site_group.test", "slug", fwresource.Slug(testName)),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_site_group", &resource.Sweeper{
		Name:         "netbox_site_group",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/site-groups/", "name")
		},
	})
}
