package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxL2VPNDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxL2VPN_basic(t *testing.T) {
	testName := testAccGetTestName("l2vpn")
	deps := testAccNetboxL2VPNDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_l2vpn", l2VPNAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_l2vpn" "test" {
  name = "%[1]s"
  type = "vxlan"
  status = "active"
  identifier = 4000
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_l2vpn.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_l2vpn.test", "type", "vxlan"),
					resource.TestCheckResourceAttr("netbox_l2vpn.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_l2vpn.test", "identifier", "4000"),
					resource.TestCheckResourceAttrSet("netbox_l2vpn.test", "slug"),
				),
			},
			{
				ResourceName:      "netbox_l2vpn.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxL2VPNDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("l2vpn_ds")
	deps := testAccNetboxL2VPNDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_l2vpn" "test" {
  name = "%[1]s"
  type = "vxlan"
  status = "active"
  identifier = 4000
}

data "netbox_l2vpn" "test" {
  depends_on = [netbox_l2vpn.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_l2vpn.test", "id", "netbox_l2vpn.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_l2vpn.test", "name", "netbox_l2vpn.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_l2vpn", &resource.Sweeper{
		Name:         "netbox_l2vpn",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "vpn/l2vpns/", "name")
		},
	})
}
