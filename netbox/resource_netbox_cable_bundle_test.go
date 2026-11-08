package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxCableBundleDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxCableBundle_basic(t *testing.T) {
	testName := testAccGetTestName("cablebundle")
	deps := testAccNetboxCableBundleDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.6.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_cable_bundle", cableBundleAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cable_bundle" "test" {
  name = "%[1]s"
  description = "test"

}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cable_bundle.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cable_bundle.test", "description", "test"),
				),
			},
			{
				ResourceName:      "netbox_cable_bundle.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCableBundleDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("cablebundle_ds")
	deps := testAccNetboxCableBundleDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.6.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cable_bundle" "test" {
  name = "%[1]s"
  description = "test"

}

data "netbox_cable_bundle" "test" {
  depends_on = [netbox_cable_bundle.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_cable_bundle.test", "id", "netbox_cable_bundle.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_cable_bundle.test", "name", "netbox_cable_bundle.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_cable_bundle", &resource.Sweeper{
		Name:         "netbox_cable_bundle",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/cable-bundles/", "name")
		},
	})
}
