package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxSavedFilterDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxSavedFilter_basic(t *testing.T) {
	testName := testAccGetTestName("savedfilter")
	deps := testAccNetboxSavedFilterDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_saved_filter", savedFilterAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_saved_filter" "test" {
  name = "%[1]s"
  object_types = ["dcim.site"]
  parameters = jsonencode({ status = ["active"] })
  weight = 10
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_saved_filter.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_saved_filter.test", "weight", "10"),
					resource.TestCheckResourceAttrSet("netbox_saved_filter.test", "slug"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_saved_filter" "test" {
  description = "updated"
  name = "%[1]s"
  object_types = ["dcim.site"]
  parameters = jsonencode({ status = ["active"] })
  weight = 10
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_saved_filter.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_saved_filter.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxSavedFilterDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("savedfilter_ds")
	deps := testAccNetboxSavedFilterDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_saved_filter" "test" {
  name = "%[1]s"
  object_types = ["dcim.site"]
  parameters = jsonencode({ status = ["active"] })
  weight = 10
}

data "netbox_saved_filter" "test" {
  depends_on = [netbox_saved_filter.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_saved_filter.test", "id", "netbox_saved_filter.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_saved_filter.test", "name", "netbox_saved_filter.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_saved_filter", &resource.Sweeper{
		Name:         "netbox_saved_filter",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "extras/saved-filters/", "name")
		},
	})
}
