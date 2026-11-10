package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxCircuitProvider_basic(t *testing.T) {
	testSlug := "circuit_prov"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_circuit_provider" "test" {
  name = "%[1]s"
  slug = "%[2]s"
  description = "This is my circuit provider!"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit_provider.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_circuit_provider.test", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_circuit_provider.test", "description", "This is my circuit provider!"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_circuit_provider" "test" {
  name = "%[1]s"
  slug = "%[2]s"
  description = "This is my circuit provider!"
}`, testName+"2", randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit_provider.test", "name", testName+"2"),
					resource.TestCheckResourceAttr("netbox_circuit_provider.test", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_circuit_provider.test", "description", "This is my circuit provider!"),
				),
			},
			{
				ResourceName:      "netbox_circuit_provider.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_circuit_provider", &resource.Sweeper{
		Name:         "netbox_circuit_provider",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "circuits/providers/", "name")
		},
	})
}
