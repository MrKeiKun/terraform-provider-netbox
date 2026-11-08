package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxDeviceModuleBaysDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("module_bays_ds")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxDeviceModuleBayFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_device_module_bay" "test" {
  device_id   = netbox_device.test.id
  name        = "%[1]s"
  label       = "%[1]s_label"
  position    = "testposition"
  description = "%[1]s_description"
  tags        = [netbox_tag.test.name]
}

data "netbox_device_module_bays" "test" {
  depends_on = [netbox_device_module_bay.test]
  filter {
    name  = "name"
    value = "%[1]s"
  }
}

data "netbox_device_module_bays" "limited" {
  depends_on = [netbox_device_module_bay.test]
  limit      = 1
  filter {
    name  = "device_id"
    value = netbox_device.test.id
  }
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_device_module_bays.test", "module_bays.#", "1"),
					resource.TestCheckResourceAttrPair("data.netbox_device_module_bays.test", "module_bays.0.id", "netbox_device_module_bay.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_device_module_bays.test", "module_bays.0.name", testName),
					resource.TestCheckResourceAttrPair("data.netbox_device_module_bays.test", "module_bays.0.device_id", "netbox_device.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_device_module_bays.test", "module_bays.0.label", testName+"_label"),
					resource.TestCheckResourceAttr("data.netbox_device_module_bays.test", "module_bays.0.position", "testposition"),
					resource.TestCheckResourceAttr("data.netbox_device_module_bays.test", "module_bays.0.description", testName+"_description"),
					resource.TestCheckResourceAttr("data.netbox_device_module_bays.test", "module_bays.0.tags.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_device_module_bays.limited", "module_bays.#", "1"),
				),
			},
		},
	})
}

func TestAccNetboxDeviceModuleBaysDataSource_empty(t *testing.T) {
	testName := testAccGetTestName("module_bays_ds_empty")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "netbox_device_module_bays" "test" {
  filter {
    name  = "name"
    value = "%s"
  }
}`, testName),
				Check: resource.TestCheckResourceAttr("data.netbox_device_module_bays.test", "module_bays.#", "0"),
			},
		},
	})
}
