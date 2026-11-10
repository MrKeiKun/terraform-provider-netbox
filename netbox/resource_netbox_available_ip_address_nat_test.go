package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxAvailableIPAddress_natInside(t *testing.T) {
	testName := testAccGetTestName("avail_ip_nat")
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_ip_address" "inside" {
  ip_address  = "10.91.1.1/24"
  status      = "active"
  description = "%[1]s"
}

resource "netbox_prefix" "test" {
  prefix = "10.91.2.0/24"
  status = "active"
}

resource "netbox_available_ip_address" "test" {
  prefix_id             = netbox_prefix.test.id
  status                = "active"
  nat_inside_address_id = netbox_ip_address.inside.id
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("netbox_available_ip_address.test", "nat_inside_address_id", "netbox_ip_address.inside", "id"),
					resource.TestCheckResourceAttr("netbox_available_ip_address.test", "ip_address", "10.91.2.1/24"),
				),
			},
			{
				ResourceName:            "netbox_available_ip_address.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"prefix_id", "ip_range_id"},
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_ip_address" "inside" {
  ip_address  = "10.91.1.1/24"
  status      = "active"
  description = "%[1]s"
}

resource "netbox_prefix" "test" {
  prefix = "10.91.2.0/24"
  status = "active"
}

resource "netbox_available_ip_address" "test" {
  prefix_id = netbox_prefix.test.id
  status    = "active"
}`, testName),
				Check: testAccCheckAttrUnset("netbox_available_ip_address.test", "nat_inside_address_id"),
			},
		},
	})
}
