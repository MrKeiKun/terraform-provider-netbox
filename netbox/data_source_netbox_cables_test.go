package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxCablesDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("cables_ds")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxCableFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_cable" "test" {
  a_termination {
    object_type = "dcim.consoleserverport"
    object_id   = netbox_device_console_server_port.test1.id
  }

  b_termination {
    object_type = "dcim.consoleport"
    object_id   = netbox_device_console_port.test1.id
  }

  status = "connected"
  label  = "%[1]s_label"
}

data "netbox_cables" "test" {
  depends_on = [netbox_cable.test]
  filter {
    name  = "label"
    value = "%[1]s_label"
  }
}

data "netbox_cables" "limited" {
  depends_on = [netbox_cable.test]
  limit      = 1
  filter {
    name  = "status"
    value = "connected"
  }
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_cables.test", "cables.#", "1"),
					resource.TestCheckResourceAttrPair("data.netbox_cables.test", "cables.0.id", "netbox_cable.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_cables.test", "cables.0.label", testName+"_label"),
					resource.TestCheckResourceAttr("data.netbox_cables.test", "cables.0.status", "connected"),
					resource.TestCheckResourceAttr("data.netbox_cables.test", "cables.0.a_terminations.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_cables.test", "cables.0.a_terminations.0.object_type", "dcim.consoleserverport"),
					resource.TestCheckResourceAttrPair("data.netbox_cables.test", "cables.0.a_terminations.0.object_id", "netbox_device_console_server_port.test1", "id"),
					resource.TestCheckResourceAttr("data.netbox_cables.test", "cables.0.b_terminations.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_cables.test", "cables.0.b_terminations.0.object_type", "dcim.consoleport"),
					resource.TestCheckResourceAttrPair("data.netbox_cables.test", "cables.0.b_terminations.0.object_id", "netbox_device_console_port.test1", "id"),
					resource.TestCheckResourceAttr("data.netbox_cables.limited", "cables.#", "1"),
				),
			},
		},
	})
}

func TestAccNetboxCablesDataSource_empty(t *testing.T) {
	testName := testAccGetTestName("cables_ds_empty")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "netbox_cables" "test" {
  filter {
    name  = "label"
    value = "%s"
  }
}`, testName),
				Check: resource.TestCheckResourceAttr("data.netbox_cables.test", "cables.#", "0"),
			},
		},
	})
}
