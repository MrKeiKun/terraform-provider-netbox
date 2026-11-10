package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxCircuitGroupAssignmentDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_circuit_provider" "test" {
  name = "%[1]s"
}

resource "netbox_circuit_type" "test" {
  name = "%[1]s"
}

resource "netbox_circuit" "test" {
  cid         = "%[1]s"
  provider_id = netbox_circuit_provider.test.id
  type_id     = netbox_circuit_type.test.id
  status      = "active"
}

resource "netbox_circuit_group" "test" {
  name = "%[1]s"
}
`, testName)
}

func TestAccNetboxCircuitGroupAssignment_basic(t *testing.T) {
	testName := testAccGetTestName("circuitgroupas")
	deps := testAccNetboxCircuitGroupAssignmentDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_circuit_group_assignment", circuitGroupAssignmentAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_circuit_group_assignment" "test" {
  group_id = netbox_circuit_group.test.id
  member_type = "circuits.circuit"
  member_id = netbox_circuit.test.id
  priority = "primary"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit_group_assignment.test", "member_type", "circuits.circuit"),
					resource.TestCheckResourceAttr("netbox_circuit_group_assignment.test", "priority", "primary"),
					resource.TestCheckResourceAttrPair("netbox_circuit_group_assignment.test", "group_id", "netbox_circuit_group.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_circuit_group_assignment" "test" {
  group_id = netbox_circuit_group.test.id
  member_type = "circuits.circuit"
  member_id = netbox_circuit.test.id
  priority = "secondary"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit_group_assignment.test", "priority", "secondary"),
				),
			},
			{
				ResourceName:      "netbox_circuit_group_assignment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
