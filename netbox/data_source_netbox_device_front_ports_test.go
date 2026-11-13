package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxDeviceFrontPortsDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("front_ports_ds")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxDeviceFrontPortFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_device_front_port" "test" {
  device_id          = netbox_device.test.id
  name               = "%[1]s"
  type               = "8p8c"
  rear_port_id       = netbox_device_rear_port.test.id
  rear_port_position = 1
  mark_connected     = true
  module_id          = netbox_module.test.id
  label              = "%[1]s_label"
  color_hex          = "123456"
  description        = "%[1]s_description"
  tags               = [netbox_tag.test.name]
}

data "netbox_device_front_ports" "test" {
  depends_on = [netbox_device_front_port.test]
  filter {
    name  = "name"
    value = "%[1]s"
  }
}

data "netbox_device_front_ports" "limited" {
  depends_on = [netbox_device_front_port.test]
  limit      = 1
  filter {
    name  = "device_id"
    value = netbox_device.test.id
  }
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_device_front_ports.test", "front_ports.#", "1"),
					resource.TestCheckResourceAttrPair("data.netbox_device_front_ports.test", "front_ports.0.id", "netbox_device_front_port.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_device_front_ports.test", "front_ports.0.name", testName),
					resource.TestCheckResourceAttrPair("data.netbox_device_front_ports.test", "front_ports.0.device_id", "netbox_device.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_device_front_ports.test", "front_ports.0.module_id", "netbox_module.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_device_front_ports.test", "front_ports.0.type", "8p8c"),
					resource.TestCheckResourceAttr("data.netbox_device_front_ports.test", "front_ports.0.label", testName+"_label"),
					resource.TestCheckResourceAttr("data.netbox_device_front_ports.test", "front_ports.0.color", "123456"),
					resource.TestCheckResourceAttr("data.netbox_device_front_ports.test", "front_ports.0.description", testName+"_description"),
					resource.TestCheckResourceAttr("data.netbox_device_front_ports.test", "front_ports.0.mark_connected", "true"),
					resource.TestCheckResourceAttr("data.netbox_device_front_ports.test", "front_ports.0.tags.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_device_front_ports.limited", "front_ports.#", "1"),
				),
			},
		},
	})
}

func TestAccNetboxDeviceFrontPortsDataSource_empty(t *testing.T) {
	testName := testAccGetTestName("front_ports_ds_empty")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "netbox_device_front_ports" "test" {
  filter {
    name  = "name"
    value = "%s"
  }
}`, testName),
				Check: resource.TestCheckResourceAttr("data.netbox_device_front_ports.test", "front_ports.#", "0"),
			},
		},
	})
}
