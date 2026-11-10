package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxVirtualCircuitTypeDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxVirtualCircuitType_basic(t *testing.T) {
	testName := testAccGetTestName("virtualcircuit")
	deps := testAccNetboxVirtualCircuitTypeDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_virtual_circuit_type", virtualCircuitTypeAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_circuit_type" "test" {
  name = "%[1]s"
  color_hex = "00ff00"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_circuit_type.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_virtual_circuit_type.test", "color_hex", "00ff00"),
					resource.TestCheckResourceAttrSet("netbox_virtual_circuit_type.test", "slug"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_circuit_type" "test" {
  description = "updated"
  name = "%[1]s"
  color_hex = "00ff00"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_circuit_type.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_virtual_circuit_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxVirtualCircuitTypeDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("virtualcircuit_ds")
	deps := testAccNetboxVirtualCircuitTypeDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_circuit_type" "test" {
  name = "%[1]s"
  color_hex = "00ff00"
}

data "netbox_virtual_circuit_type" "test" {
  depends_on = [netbox_virtual_circuit_type.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_virtual_circuit_type.test", "id", "netbox_virtual_circuit_type.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_virtual_circuit_type.test", "name", "netbox_virtual_circuit_type.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_virtual_circuit_type", &resource.Sweeper{
		Name:         "netbox_virtual_circuit_type",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "circuits/virtual-circuit-types/", "name")
		},
	})
}
