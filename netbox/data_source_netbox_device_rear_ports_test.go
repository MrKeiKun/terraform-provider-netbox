package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxDeviceRearPortsDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("rear_ports_ds")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxDeviceFrontPortFullDependencies(testName) + fmt.Sprintf(`
data "netbox_device_rear_ports" "test" {
  depends_on = [netbox_device_rear_port.test]
  filter {
    name  = "name"
    value = "%[1]s"
  }
}

data "netbox_device_rear_ports" "limited" {
  depends_on = [netbox_device_rear_port.test]
  limit      = 1
  filter {
    name  = "device_id"
    value = netbox_device.test.id
  }
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_device_rear_ports.test", "rear_ports.#", "1"),
					resource.TestCheckResourceAttrPair("data.netbox_device_rear_ports.test", "rear_ports.0.id", "netbox_device_rear_port.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_device_rear_ports.test", "rear_ports.0.name", testName),
					resource.TestCheckResourceAttrPair("data.netbox_device_rear_ports.test", "rear_ports.0.device_id", "netbox_device.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_device_rear_ports.test", "rear_ports.0.type", "8p8c"),
					resource.TestCheckResourceAttr("data.netbox_device_rear_ports.test", "rear_ports.0.positions", "1"),
					resource.TestCheckResourceAttr("data.netbox_device_rear_ports.test", "rear_ports.0.mark_connected", "true"),
					resource.TestCheckResourceAttr("data.netbox_device_rear_ports.limited", "rear_ports.#", "1"),
				),
			},
		},
	})
}

func TestAccNetboxDeviceRearPortsDataSource_empty(t *testing.T) {
	testName := testAccGetTestName("rear_ports_ds_empty")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "netbox_device_rear_ports" "test" {
  filter {
    name  = "name"
    value = "%s"
  }
}`, testName),
				Check: resource.TestCheckResourceAttr("data.netbox_device_rear_ports.test", "rear_ports.#", "0"),
			},
		},
	})
}
