package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxWirelessLinkDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = "%[1]s"
  status = "active"
}

resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_device_type" "test" {
  model           = "%[1]s"
  manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_device_role" "test" {
  name      = "%[1]s"
  color_hex = "123456"
}

resource "netbox_device" "test" {
  name           = "%[1]s"
  device_type_id = netbox_device_type.test.id
  role_id        = netbox_device_role.test.id
  site_id        = netbox_site.test.id
}

resource "netbox_device_interface" "a" {
  name      = "%[1]s-a"
  device_id = netbox_device.test.id
  type      = "ieee802.11ac"
}

resource "netbox_device_interface" "b" {
  name      = "%[1]s-b"
  device_id = netbox_device.test.id
  type      = "ieee802.11ac"
}
`, testName)
}

func TestAccNetboxWirelessLink_basic(t *testing.T) {
	testName := testAccGetTestName("wirelesslink")
	deps := testAccNetboxWirelessLinkDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_wireless_link", wirelessLinkAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_wireless_link" "test" {
  interface_a_id = netbox_device_interface.a.id
  interface_b_id = netbox_device_interface.b.id
  ssid = "%[1]s"
  status = "connected"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_wireless_link.test", "ssid", testName),
					resource.TestCheckResourceAttr("netbox_wireless_link.test", "status", "connected"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_wireless_link" "test" {
  description = "updated"
  interface_a_id = netbox_device_interface.a.id
  interface_b_id = netbox_device_interface.b.id
  ssid = "%[1]s"
  status = "connected"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_wireless_link.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_wireless_link.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxWirelessLinkDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("wirelesslink_ds")
	deps := testAccNetboxWirelessLinkDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_wireless_link" "test" {
  interface_a_id = netbox_device_interface.a.id
  interface_b_id = netbox_device_interface.b.id
  ssid = "%[1]s"
  status = "connected"
}

data "netbox_wireless_link" "test" {
  depends_on = [netbox_wireless_link.test]
  ssid = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_wireless_link.test", "id", "netbox_wireless_link.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_wireless_link.test", "ssid", "netbox_wireless_link.test", "ssid"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_wireless_link", &resource.Sweeper{
		Name:         "netbox_wireless_link",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "wireless/wireless-links/", "ssid")
		},
	})
}
