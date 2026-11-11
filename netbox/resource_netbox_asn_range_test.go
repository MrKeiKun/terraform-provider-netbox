package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxASNRangeDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_rir" "test" {
  name = "%[1]s"
}
`, testName)
}

func TestAccNetboxASNRange_basic(t *testing.T) {
	testName := testAccGetTestName("asnrange")
	deps := testAccNetboxASNRangeDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_asn_range", "ipam/asn-ranges/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_asn_range" "test" {
  name = "%[1]s"
  rir_id = netbox_rir.test.id
  start = 65000
  end = 65100
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_asn_range.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_asn_range.test", "start", "65000"),
					resource.TestCheckResourceAttr("netbox_asn_range.test", "end", "65100"),
					resource.TestCheckResourceAttrSet("netbox_asn_range.test", "slug"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_asn_range" "test" {
  description = "updated"
  name = "%[1]s"
  rir_id = netbox_rir.test.id
  start = 65000
  end = 65100
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_asn_range.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_asn_range.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxASNRangeDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("asnrange_ds")
	deps := testAccNetboxASNRangeDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_asn_range" "test" {
  name = "%[1]s"
  rir_id = netbox_rir.test.id
  start = 65000
  end = 65100
}

data "netbox_asn_range" "test" {
  depends_on = [netbox_asn_range.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_asn_range.test", "id", "netbox_asn_range.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_asn_range.test", "name", "netbox_asn_range.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_asn_range", &resource.Sweeper{
		Name:         "netbox_asn_range",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "ipam/asn-ranges/", "name")
		},
	})
}
