package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxAggregate_basic(t *testing.T) {
	testPrefix := "1.1.1.0/25"
	testSlug := "aggregate"
	testDesc := "test aggregate"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_rir" "test" {
  name = "%s"
  slug = "%s"
}
resource "netbox_aggregate" "test" {
  prefix = "%s"
  description = "%s"
  rir_id = netbox_rir.test.id
}`, testName, randomSlug, testPrefix, testDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_aggregate.test", "prefix", testPrefix),
					resource.TestCheckResourceAttr("netbox_aggregate.test", "description", testDesc),
					resource.TestCheckResourceAttrPair("netbox_aggregate.test", "rir_id", "netbox_rir.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_rir" "test" {
  name = "%s"
  slug = "%s"
}
resource "netbox_aggregate" "test" {
  prefix = "%s"
  description = "%s-updated"
  rir_id = netbox_rir.test.id
}`, testName, randomSlug, testPrefix, testDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_aggregate.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_aggregate.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_aggregate", &resource.Sweeper{
		Name:         "netbox_aggregate",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "ipam/aggregates/", "prefix")
		},
	})
}
