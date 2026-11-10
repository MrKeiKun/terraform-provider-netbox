package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxVpnTunnelGroup_basic(t *testing.T) {
	testSlug := "vpntnlgrp_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_vpn_tunnel_group" "test" {
  name = "%[1]s"
  slug = "%[1]s"
  description = "%[1]s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vpn_tunnel_group.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_vpn_tunnel_group.test", "slug", testName),
					resource.TestCheckResourceAttr("netbox_vpn_tunnel_group.test", "description", testName),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_vpn_tunnel_group" "test" {
  name = "%[1]s"
  slug = "%[1]s"
  description = "%[1]s-updated"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_vpn_tunnel_group.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_vpn_tunnel_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_vpn_tunnel_group", &resource.Sweeper{
		Name:         "netbox_vpn_tunnel_group",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "vpn/tunnel-groups/", "name")
		},
	})
}
