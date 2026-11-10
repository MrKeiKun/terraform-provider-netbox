package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxVirtualDeviceContextDependencies(testName string) string {
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
`, testName)
}

func TestAccNetboxVirtualDeviceContext_basic(t *testing.T) {
	testName := testAccGetTestName("virtualdevicec")
	deps := testAccNetboxVirtualDeviceContextDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_virtual_device_context", virtualDeviceContextAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_device_context" "test" {
  name = "%[1]s"
  device_id = netbox_device.test.id
  status = "active"
  identifier = 7
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_device_context.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_virtual_device_context.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_virtual_device_context.test", "identifier", "7"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_device_context" "test" {
  description = "updated"
  name = "%[1]s"
  device_id = netbox_device.test.id
  status = "active"
  identifier = 7
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_device_context.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_virtual_device_context.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxVirtualDeviceContextDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("virtualdevicec_ds")
	deps := testAccNetboxVirtualDeviceContextDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_device_context" "test" {
  name = "%[1]s"
  device_id = netbox_device.test.id
  status = "active"
  identifier = 7
}

data "netbox_virtual_device_context" "test" {
  depends_on = [netbox_virtual_device_context.test]
  name = "%[1]s"
  device_id = netbox_device.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_virtual_device_context.test", "id", "netbox_virtual_device_context.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_virtual_device_context.test", "name", "netbox_virtual_device_context.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_virtual_device_context", &resource.Sweeper{
		Name:         "netbox_virtual_device_context",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/virtual-device-contexts/", "name")
		},
	})
}
