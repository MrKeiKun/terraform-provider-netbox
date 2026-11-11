package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxPowerPanelFullDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_site" "test" {
  name = "%[1]s"
  status = "active"
}

resource "netbox_location" "test" {
  name = "%[1]s"
  site_id =netbox_site.test.id
}

resource "netbox_tag" "test_a" {
  name = "%[1]sa"
}`, testName)
}

func TestAccNetboxPowerPanel_basic(t *testing.T) {
	testSlug := "power_panel_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckPowerPanelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxPowerPanelFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_power_panel" "test" {
  name = "%[1]s"
  description = "%[1]sdescription"
  comments = "%[1]scomments"

  site_id = netbox_site.test.id
  location_id = netbox_location.test.id
  tags = [netbox_tag.test_a.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_power_panel.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_power_panel.test", "description", testName+"description"),
					resource.TestCheckResourceAttr("netbox_power_panel.test", "comments", testName+"comments"),
					resource.TestCheckResourceAttr("netbox_power_panel.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_power_panel.test", "tags.0", testName+"a"),

					resource.TestCheckResourceAttrPair("netbox_power_panel.test", "site_id", "netbox_site.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_power_panel.test", "location_id", "netbox_location.test", "id"),
				),
			},
			{
				Config: testAccNetboxPowerPanelFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_power_panel" "test" {
  name = "%[1]s"
  site_id = netbox_site.test.id
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_power_panel.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_power_panel.test", "description", ""),
					resource.TestCheckResourceAttr("netbox_power_panel.test", "comments", ""),
					resource.TestCheckResourceAttr("netbox_power_panel.test", "tags.#", "0"),
					resource.TestCheckResourceAttr("netbox_power_panel.test", "location_id", "0"),

					resource.TestCheckResourceAttrPair("netbox_power_panel.test", "site_id", "netbox_site.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_power_panel.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPowerPanelDestroy(s *terraform.State) error {
	return testAccCheckObjectsDestroyed(s, "netbox_power_panel", "dcim/power-panels/")
}

func init() {
	resource.AddTestSweepers("netbox_power_panel", &resource.Sweeper{
		Name:         "netbox_power_panel",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/power-panels/", "name")
		},
	})
}
