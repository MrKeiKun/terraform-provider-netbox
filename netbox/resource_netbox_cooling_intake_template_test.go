package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxCoolingIntakeTemplateDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name = "%[1]s"
}

resource "netbox_device_type" "test" {
  model           = "%[1]s"
  manufacturer_id = netbox_manufacturer.test.id
}
`, testName)
}

func TestAccNetboxCoolingIntakeTemplate_basic(t *testing.T) {
	testName := testAccGetTestName("cintaketpl")
	deps := testAccNetboxCoolingIntakeTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_cooling_intake_template", "dcim/cooling-intake-templates/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_intake_template" "test" {
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type = "uqd"
  diameter = 12.5
  diameter_unit = "mm"
  max_flow = 20
  max_flow_unit = "lpm"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_intake_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cooling_intake_template.test", "type", "uqd"),
					resource.TestCheckResourceAttr("netbox_cooling_intake_template.test", "diameter", "12.5"),
					resource.TestCheckResourceAttr("netbox_cooling_intake_template.test", "diameter_unit", "mm"),
					resource.TestCheckResourceAttr("netbox_cooling_intake_template.test", "max_flow", "20"),
					resource.TestCheckResourceAttr("netbox_cooling_intake_template.test", "max_flow_unit", "lpm"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_intake_template" "test" {
  description = "updated"
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type = "uqd"
  diameter = 12.5
  diameter_unit = "mm"
  max_flow = 20
  max_flow_unit = "lpm"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_intake_template.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_cooling_intake_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCoolingIntakeTemplateDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("cintaketpl_ds")
	deps := testAccNetboxCoolingIntakeTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_intake_template" "test" {
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type = "uqd"
  diameter = 12.5
  diameter_unit = "mm"
  max_flow = 20
  max_flow_unit = "lpm"
}

data "netbox_cooling_intake_template" "test" {
  depends_on = [netbox_cooling_intake_template.test]
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_cooling_intake_template.test", "id", "netbox_cooling_intake_template.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_cooling_intake_template.test", "name", "netbox_cooling_intake_template.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_cooling_intake_template", &resource.Sweeper{
		Name:         "netbox_cooling_intake_template",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/cooling-intake-templates/", "name")
		},
	})
}
