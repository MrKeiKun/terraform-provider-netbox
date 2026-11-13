package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxVirtualCircuitTerminationDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_circuit_provider" "test" {
  name = "%[1]s"
}

resource "netbox_circuit_provider_network" "test" {
  name        = "%[1]s"
  provider_id = netbox_circuit_provider.test.id
}

resource "netbox_virtual_circuit_type" "test" {
  name = "%[1]s"
}

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

resource "netbox_device_interface" "test" {
  name      = "%[1]s"
  device_id = netbox_device.test.id
  type      = "virtual"
}

resource "netbox_virtual_circuit" "test" {
  cid                 = "%[1]s"
  provider_network_id = netbox_circuit_provider_network.test.id
  type_id             = netbox_virtual_circuit_type.test.id
}
`, testName)
}

func TestAccNetboxVirtualCircuitTermination_basic(t *testing.T) {
	testName := testAccGetTestName("virtualcircuit")
	deps := testAccNetboxVirtualCircuitTerminationDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_virtual_circuit_termination", "circuits/virtual-circuit-terminations/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_circuit_termination" "test" {
  virtual_circuit_id = netbox_virtual_circuit.test.id
  interface_id = netbox_device_interface.test.id
  role = "peer"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_circuit_termination.test", "role", "peer"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_circuit_termination" "test" {
  description = "updated"
  virtual_circuit_id = netbox_virtual_circuit.test.id
  interface_id = netbox_device_interface.test.id
  role = "peer"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_circuit_termination.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_virtual_circuit_termination.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
