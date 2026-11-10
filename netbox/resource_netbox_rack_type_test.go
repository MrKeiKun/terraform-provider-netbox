package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxRackTypeFullDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_tag" "test_a" {
  name = "%[1]sa"
}`, testName)
}

func TestAccNetboxRackType_basic(t *testing.T) {
	testSlug := "racktype_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRackTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxRackTypeFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_rack_type" "test" {
  model             = "%[1]s"
  manufacturer_id   = netbox_manufacturer.test.id
  width             = 19
  u_height          = 48
  starting_unit     = 1
  form_factor       = "2-post-frame"
  tags              = [netbox_tag.test_a.name]
  description       = "%[1]s"
  outer_width       = 10
  outer_depth       = 15
  outer_unit        = "mm"
  weight            = 15
  max_weight        = 20
  weight_unit       = "kg"
  mounting_depth_mm = 21
  comments          = "%[1]scomments"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rack_type.test", "model", testName),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "description", testName),
					resource.TestCheckResourceAttrPair("netbox_rack_type.test", "manufacturer_id", "netbox_manufacturer.test", "id"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "width", "19"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "u_height", "48"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "tags.0", testName+"a"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "outer_width", "10"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "outer_depth", "15"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "outer_unit", "mm"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "weight", "15"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "max_weight", "20"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "weight_unit", "kg"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "comments", testName+"comments"),
					resource.TestCheckResourceAttr("netbox_rack_type.test", "mounting_depth_mm", "21"),
				),
			},
			{
				Config: testAccNetboxRackTypeFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_rack_type" "test" {
  model             = "%[1]s"
  manufacturer_id   = netbox_manufacturer.test.id
  width             = 19
  u_height          = 48
  starting_unit     = 1
  form_factor       = "2-post-frame"
  tags              = [netbox_tag.test_a.name]
  description       = "%[1]s-updated"
  outer_width       = 10
  outer_depth       = 15
  outer_unit        = "mm"
  weight            = 15
  max_weight        = 20
  weight_unit       = "kg"
  mounting_depth_mm = 21
  comments          = "%[1]scomments"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_rack_type.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_rack_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRackTypeDestroy(s *terraform.State) error {
	return testAccCheckObjectsDestroyed(s, "netbox_rack_type", "dcim/rack-types/")
}

func init() {
	resource.AddTestSweepers("netbox_rack_type", &resource.Sweeper{
		Name:         "netbox_rack_type",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/rack-types/", "model")
		},
	})
}
