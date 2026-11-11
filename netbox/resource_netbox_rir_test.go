package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxRir_basic(t *testing.T) {
	testSlug := "rir"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_rir" "test_basic" {
  name        = "%s"
  slug        = "%s"
  description = "my-description"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rir.test_basic", "name", testName),
					resource.TestCheckResourceAttr("netbox_rir.test_basic", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_rir.test_basic", "description", "my-description"),
				),
			},
			{
				ResourceName:      "netbox_rir.test_basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxRir_privacy(t *testing.T) {
	testSlug := "rir_privacy"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_rir" "test_privacy" {
  name        = "%s"
  slug        = "%s"
  is_private  = false
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rir.test_privacy", "name", testName),
					resource.TestCheckResourceAttr("netbox_rir.test_privacy", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_rir.test_privacy", "is_private", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_rir" "test_privacy" {
  name        = "%s"
  slug        = "%s"
  is_private  = true
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rir.test_privacy", "name", testName),
					resource.TestCheckResourceAttr("netbox_rir.test_privacy", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_rir.test_privacy", "is_private", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_rir" "test_privacy" {
  name        = "%s"
  slug        = "%s"
  is_private  = false
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rir.test_privacy", "name", testName),
					resource.TestCheckResourceAttr("netbox_rir.test_privacy", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_rir.test_privacy", "is_private", "false"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_rir", &resource.Sweeper{
		Name:         "netbox_rir",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "ipam/rirs/", "name")
		},
	})
}
