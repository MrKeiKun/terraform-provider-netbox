package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxL2VPNTerminationDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_l2vpn" "test" {
  name = "%[1]s"
  type = "vxlan"
}

resource "netbox_vlan" "test" {
  name = "%[1]s"
  vid  = 3777
}
`, testName)
}

func TestAccNetboxL2VPNTermination_basic(t *testing.T) {
	testName := testAccGetTestName("l2vpnterminati")
	deps := testAccNetboxL2VPNTerminationDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_l2vpn_termination", l2VPNTerminationAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_l2vpn_termination" "test" {
  l2vpn_id = netbox_l2vpn.test.id
  assigned_object_type = "ipam.vlan"
  assigned_object_id = netbox_vlan.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_l2vpn_termination.test", "assigned_object_type", "ipam.vlan"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_vlan" "test2" {
  name = "%[1]s-2"
  vid  = 3778
}
resource "netbox_l2vpn_termination" "test" {
  l2vpn_id = netbox_l2vpn.test.id
  assigned_object_type = "ipam.vlan"
  assigned_object_id = netbox_vlan.test2.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("netbox_l2vpn_termination.test", "assigned_object_id", "netbox_vlan.test2", "id"),
				),
			},
			{
				ResourceName:      "netbox_l2vpn_termination.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
