package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxCircuitType_basic(t *testing.T) {
	testSlug := "circuit_type"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_circuit_type" "test" {
  name = "%[1]s"
  slug = "%[2]s"
  description = "This is my circuit type!"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit_type.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_circuit_type.test", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_circuit_type.test", "description", "This is my circuit type!"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_circuit_type" "test" {
  name = "%[1]s"
  slug = "%[2]s"
  description = "This is my circuit type!-updated"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_circuit_type.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_circuit_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_circuit_type", &resource.Sweeper{
		Name:         "netbox_circuit_type",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "circuits/circuit-types/", "name")
		},
	})
}
