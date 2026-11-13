package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxCoolingSourceDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = "%[1]s"
  status = "active"
}
`, testName)
}

func TestAccNetboxCoolingSource_basic(t *testing.T) {
	testName := testAccGetTestName("coolingsource")
	deps := testAccNetboxCoolingSourceDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_cooling_source", "dcim/cooling-sources/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_source" "test" {
  name = "%[1]s"
  site_id = netbox_site.test.id
  type = "chiller"
  status = "active"
  cooling_capacity = 100

}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_source.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cooling_source.test", "type", "chiller"),
					resource.TestCheckResourceAttr("netbox_cooling_source.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_cooling_source.test", "cooling_capacity", "100"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_source" "test" {
  description = "updated"
  name = "%[1]s"
  site_id = netbox_site.test.id
  type = "chiller"
  status = "active"
  cooling_capacity = 100

}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_source.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_cooling_source.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCoolingSourceDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("coolingsource_ds")
	deps := testAccNetboxCoolingSourceDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_source" "test" {
  name = "%[1]s"
  site_id = netbox_site.test.id
  type = "chiller"
  status = "active"
  cooling_capacity = 100

}

data "netbox_cooling_source" "test" {
  depends_on = [netbox_cooling_source.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_cooling_source.test", "id", "netbox_cooling_source.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_cooling_source.test", "name", "netbox_cooling_source.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_cooling_source", &resource.Sweeper{
		Name:         "netbox_cooling_source",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/cooling-sources/", "name")
		},
	})
}
