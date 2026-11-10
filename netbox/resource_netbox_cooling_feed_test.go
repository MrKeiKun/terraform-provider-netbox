package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxCoolingFeedDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = "%[1]s"
  status = "active"
}

resource "netbox_cooling_source" "test" {
  name    = "%[1]s"
  site_id = netbox_site.test.id
  type    = "chiller"
}
`, testName)
}

func TestAccNetboxCoolingFeed_basic(t *testing.T) {
	testName := testAccGetTestName("coolingfeed")
	deps := testAccNetboxCoolingFeedDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_cooling_feed", coolingFeedAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_feed" "test" {
  name = "%[1]s"
  cooling_source_id = netbox_cooling_source.test.id
  status = "active"
  max_flow = 10.5
  max_flow_unit = "lpm"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_feed.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cooling_feed.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_cooling_feed.test", "max_flow", "10.5"),
					resource.TestCheckResourceAttr("netbox_cooling_feed.test", "max_flow_unit", "lpm"),
					resource.TestCheckResourceAttrPair("netbox_cooling_feed.test", "cooling_source_id", "netbox_cooling_source.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_feed" "test" {
  description = "updated"
  name = "%[1]s"
  cooling_source_id = netbox_cooling_source.test.id
  status = "active"
  max_flow = 10.5
  max_flow_unit = "lpm"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_feed.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_cooling_feed.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCoolingFeedDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("coolingfeed_ds")
	deps := testAccNetboxCoolingFeedDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_feed" "test" {
  name = "%[1]s"
  cooling_source_id = netbox_cooling_source.test.id
  status = "active"
  max_flow = 10.5
  max_flow_unit = "lpm"
}

data "netbox_cooling_feed" "test" {
  depends_on = [netbox_cooling_feed.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_cooling_feed.test", "id", "netbox_cooling_feed.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_cooling_feed.test", "name", "netbox_cooling_feed.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_cooling_feed", &resource.Sweeper{
		Name:         "netbox_cooling_feed",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/cooling-feeds/", "name")
		},
	})
}
