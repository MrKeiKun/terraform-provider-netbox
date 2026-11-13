package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxIpAddressDataSource_basic(t *testing.T) {
	ipAddress := "10.0.0.107/24"
	status := "active"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_ip_address" "test" {
  ip_address = "%[1]s"
  status = "%[2]s"
}
data "netbox_ip_address" "test" {
  depends_on = [netbox_ip_address.test]
  id = netbox_ip_address.test.id
}`, ipAddress, status),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_ip_address.test", "ip_address", "netbox_ip_address.test", "ip_address"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
