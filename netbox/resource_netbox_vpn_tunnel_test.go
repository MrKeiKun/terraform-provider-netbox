package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxVpnTunnel_basic(t *testing.T) {
	testSlug := "vpntun_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_vpn_tunnel_group" "test" {
  name = "%[1]s"
  description = "%[1]s"
}
resource "netbox_tenant" "test" {
  name = "%[1]s"
}
resource "netbox_tag" "test" {
  name = "%[1]s"
}
resource "netbox_vpn_tunnel" "test" {
  name = "%[1]s"
  encapsulation = "ipsec-transport"
  status = "active"
  tunnel_group_id = netbox_vpn_tunnel_group.test.id

  description = "%[1]s"
  tenant_id = netbox_tenant.test.id
  tunnel_id = 123
  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vpn_tunnel.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_vpn_tunnel.test", "encapsulation", "ipsec-transport"),
					resource.TestCheckResourceAttr("netbox_vpn_tunnel.test", "description", testName),
					resource.TestCheckResourceAttr("netbox_vpn_tunnel.test", "tunnel_id", "123"),
					resource.TestCheckResourceAttrPair("netbox_vpn_tunnel.test", "tenant_id", "netbox_tenant.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_vpn_tunnel.test", "tunnel_group_id", "netbox_vpn_tunnel_group.test", "id"),
					resource.TestCheckResourceAttr("netbox_vpn_tunnel.test", "tags.0", testName),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_vpn_tunnel_group" "test" {
  name = "%[1]s"
  description = "%[1]s"
}
resource "netbox_tenant" "test" {
  name = "%[1]s"
}
resource "netbox_tag" "test" {
  name = "%[1]s"
}
resource "netbox_vpn_tunnel" "test" {
  name = "%[1]s"
  encapsulation = "ipsec-transport"
  status = "active"
  tunnel_group_id = netbox_vpn_tunnel_group.test.id

  description = "%[1]s-updated"
  tenant_id = netbox_tenant.test.id
  tunnel_id = 123
  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_vpn_tunnel.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_vpn_tunnel.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_vpn_tunnel", &resource.Sweeper{
		Name:         "netbox_vpn_tunnel",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "vpn/tunnels/", "name")
		},
	})
}
