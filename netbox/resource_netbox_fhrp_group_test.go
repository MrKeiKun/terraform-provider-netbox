package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxFhrpGroup_basic(t *testing.T) {
	testSlug := "fhrp_group_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`

resource "netbox_tag" "test" {
  name = "%[1]sa"
}

resource "netbox_fhrp_group" "test" {
  protocol    = "other"
  group_id    = "1234"
  auth_type   = "md5"
  auth_key    = "test"
  name        = "test"
  description = "test"
  comments    = "test"

  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_fhrp_group.test", "protocol", "other"),
					resource.TestCheckResourceAttr("netbox_fhrp_group.test", "group_id", "1234"),
					resource.TestCheckResourceAttr("netbox_fhrp_group.test", "auth_type", "md5"),
					resource.TestCheckResourceAttr("netbox_fhrp_group.test", "auth_key", "test"),
					resource.TestCheckResourceAttr("netbox_fhrp_group.test", "name", "test"),
					resource.TestCheckResourceAttr("netbox_fhrp_group.test", "description", "test"),
					resource.TestCheckResourceAttr("netbox_fhrp_group.test", "comments", "test"),
					resource.TestCheckResourceAttr("netbox_fhrp_group.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_fhrp_group.test", "tags.0", testName+"a"),
				),
			},
			{
				Config: fmt.Sprintf(`

resource "netbox_tag" "test" {
  name = "%[1]sa"
}

resource "netbox_fhrp_group" "test" {
  protocol    = "other"
  group_id    = "1234"
  auth_type   = "md5"
  auth_key    = "test"
  name        = "test"
  description = "test-updated"
  comments    = "test"

  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_fhrp_group.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_fhrp_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_fhrp_group", &resource.Sweeper{
		Name:         "netbox_fhrp_group",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "ipam/fhrp-groups/", "name")
		},
	})
}
