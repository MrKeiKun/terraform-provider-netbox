package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccNetboxCircuitDependencies(testName string, testSlug string) string {
	return fmt.Sprintf(`
resource "netbox_tenant" "test" {
	name = "%[1]s"
	slug = "%[2]s"
}
resource "netbox_circuit_provider" "test" {
	name = "%[1]s"
	slug = "%[2]s"
}
resource "netbox_circuit_type" "test" {
	name = "%[1]s"
	slug = "%[2]s"
}
`, testName, testSlug)
}
func TestAccNetboxCircuit_basic(t *testing.T) {
	testSlug := "circuit_prov"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxCircuitDependencies(testName, randomSlug) + fmt.Sprintf(`
resource "netbox_circuit" "test" {
  cid = "%[1]s"
  status = "active"
  description = "This is my circuit!"
  provider_id = netbox_circuit_provider.test.id
  type_id = netbox_circuit_type.test.id
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit.test", "cid", testName),
					resource.TestCheckResourceAttrPair("netbox_circuit.test", "provider_id", "netbox_circuit_provider.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_circuit.test", "type_id", "netbox_circuit_type.test", "id"),
					resource.TestCheckResourceAttr("netbox_circuit.test", "description", "This is my circuit!"),
				),
			},
			{
				Config: testAccNetboxCircuitDependencies(testName, randomSlug) + fmt.Sprintf(`
resource "netbox_circuit" "test" {
  cid = "%[1]s"
  status = "active"
  description = "This is my circuit!"
  provider_id = netbox_circuit_provider.test.id
  type_id = netbox_circuit_type.test.id
  tenant_id = netbox_tenant.test.id
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit.test", "cid", testName),
					resource.TestCheckResourceAttrPair("netbox_circuit.test", "provider_id", "netbox_circuit_provider.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_circuit.test", "type_id", "netbox_circuit_type.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_circuit.test", "tenant_id", "netbox_tenant.test", "id"),
					resource.TestCheckResourceAttr("netbox_circuit.test", "description", "This is my circuit!"),
				),
			},
			{
				ResourceName:      "netbox_circuit.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_circuit", &resource.Sweeper{
		Name:         "netbox_circuit",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "circuits/circuits/", "cid")
		},
	})
}
