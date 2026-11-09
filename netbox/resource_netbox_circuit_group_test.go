package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxCircuitGroupDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxCircuitGroup_basic(t *testing.T) {
	testName := testAccGetTestName("circuitgroup")
	deps := testAccNetboxCircuitGroupDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_circuit_group", circuitGroupAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_circuit_group" "test" {
  name = "%[1]s"
  description = "test"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit_group.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_circuit_group.test", "description", "test"),
					resource.TestCheckResourceAttrSet("netbox_circuit_group.test", "slug"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_circuit_group" "test" {
  name = "%[1]s"
  description = "test-updated"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_circuit_group.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_circuit_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCircuitGroupDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("circuitgroup_ds")
	deps := testAccNetboxCircuitGroupDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_circuit_group" "test" {
  name = "%[1]s"
  description = "test"
}

data "netbox_circuit_group" "test" {
  depends_on = [netbox_circuit_group.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_circuit_group.test", "id", "netbox_circuit_group.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_circuit_group.test", "name", "netbox_circuit_group.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_circuit_group", &resource.Sweeper{
		Name:         "netbox_circuit_group",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "circuits/circuit-groups/", "name")
		},
	})
}
