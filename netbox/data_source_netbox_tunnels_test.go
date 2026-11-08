package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxTunnelsDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("tunnels_ds")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_vpn_tunnel_group" "test" {
  name = "%[1]s"
}

resource "netbox_tenant" "test" {
  name = "%[1]s"
}

resource "netbox_tag" "test" {
  name = "%[1]s"
}

resource "netbox_vpn_tunnel" "test" {
  name            = "%[1]s"
  encapsulation   = "ipsec-transport"
  status          = "active"
  tunnel_group_id = netbox_vpn_tunnel_group.test.id
  description     = "%[1]s_description"
  tenant_id       = netbox_tenant.test.id
  tunnel_id       = 123
  tags            = [netbox_tag.test.name]
}

data "netbox_tunnels" "test" {
  depends_on = [netbox_vpn_tunnel.test]
  filter {
    name  = "name"
    value = "%[1]s"
  }
}

data "netbox_tunnels" "limited" {
  depends_on = [netbox_vpn_tunnel.test]
  limit      = 1
  filter {
    name  = "tenant_id"
    value = netbox_tenant.test.id
  }
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_tunnels.test", "tunnels.#", "1"),
					resource.TestCheckResourceAttrPair("data.netbox_tunnels.test", "tunnels.0.id", "netbox_vpn_tunnel.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_tunnels.test", "tunnels.0.name", testName),
					resource.TestCheckResourceAttr("data.netbox_tunnels.test", "tunnels.0.status", "active"),
					resource.TestCheckResourceAttr("data.netbox_tunnels.test", "tunnels.0.encapsulation", "ipsec-transport"),
					resource.TestCheckResourceAttrPair("data.netbox_tunnels.test", "tunnels.0.group_id", "netbox_vpn_tunnel_group.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_tunnels.test", "tunnels.0.tenant_id", "netbox_tenant.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_tunnels.test", "tunnels.0.tunnel_id", "123"),
					resource.TestCheckResourceAttr("data.netbox_tunnels.test", "tunnels.0.description", testName+"_description"),
					resource.TestCheckResourceAttr("data.netbox_tunnels.test", "tunnels.0.tags.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_tunnels.limited", "tunnels.#", "1"),
				),
			},
		},
	})
}

func TestAccNetboxTunnelsDataSource_empty(t *testing.T) {
	testName := testAccGetTestName("tunnels_ds_empty")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "netbox_tunnels" "test" {
  filter {
    name  = "name"
    value = "%s"
  }
}`, testName),
				Check: resource.TestCheckResourceAttr("data.netbox_tunnels.test", "tunnels.#", "0"),
			},
		},
	})
}
