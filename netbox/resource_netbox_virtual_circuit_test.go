package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxVirtualCircuitDependencies(testName string) string {
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
`, testName)
}

func TestAccNetboxVirtualCircuit_basic(t *testing.T) {
	testName := testAccGetTestName("virtualcircuit")
	deps := testAccNetboxVirtualCircuitDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_virtual_circuit", virtualCircuitAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_circuit" "test" {
  cid = "%[1]s"
  provider_network_id = netbox_circuit_provider_network.test.id
  type_id = netbox_virtual_circuit_type.test.id
  status = "active"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_circuit.test", "cid", testName),
					resource.TestCheckResourceAttr("netbox_virtual_circuit.test", "status", "active"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_circuit" "test" {
  description = "updated"
  cid = "%[1]s"
  provider_network_id = netbox_circuit_provider_network.test.id
  type_id = netbox_virtual_circuit_type.test.id
  status = "active"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_circuit.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_virtual_circuit.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxVirtualCircuitDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("virtualcircuit_ds")
	deps := testAccNetboxVirtualCircuitDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_circuit" "test" {
  cid = "%[1]s"
  provider_network_id = netbox_circuit_provider_network.test.id
  type_id = netbox_virtual_circuit_type.test.id
  status = "active"
}

data "netbox_virtual_circuit" "test" {
  depends_on = [netbox_virtual_circuit.test]
  cid = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_virtual_circuit.test", "id", "netbox_virtual_circuit.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_virtual_circuit.test", "cid", "netbox_virtual_circuit.test", "cid"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_virtual_circuit", &resource.Sweeper{
		Name:         "netbox_virtual_circuit",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "circuits/virtual-circuits/", "cid")
		},
	})
}
