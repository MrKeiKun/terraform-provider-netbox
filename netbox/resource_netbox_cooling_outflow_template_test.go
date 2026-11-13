package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxCoolingOutflowTemplateDependencies(testName string) string {
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

func TestAccNetboxCoolingOutflowTemplate_basic(t *testing.T) {
	testName := testAccGetTestName("coutflowtpl")
	deps := testAccNetboxCoolingOutflowTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_cooling_outflow_template", "dcim/cooling-outflow-templates/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_outflow_template" "test" {
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type = "uqd"
  diameter = 12.5
  diameter_unit = "mm"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_outflow_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cooling_outflow_template.test", "type", "uqd"),
					resource.TestCheckResourceAttr("netbox_cooling_outflow_template.test", "diameter", "12.5"),
					resource.TestCheckResourceAttr("netbox_cooling_outflow_template.test", "diameter_unit", "mm"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_outflow_template" "test" {
  description = "updated"
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type = "uqd"
  diameter = 12.5
  diameter_unit = "mm"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cooling_outflow_template.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_cooling_outflow_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCoolingOutflowTemplateDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("coutflowtpl_ds")
	deps := testAccNetboxCoolingOutflowTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.7.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_cooling_outflow_template" "test" {
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
  type = "uqd"
  diameter = 12.5
  diameter_unit = "mm"
}

data "netbox_cooling_outflow_template" "test" {
  depends_on = [netbox_cooling_outflow_template.test]
  name = "%[1]s"
  device_type_id = netbox_device_type.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_cooling_outflow_template.test", "id", "netbox_cooling_outflow_template.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_cooling_outflow_template.test", "name", "netbox_cooling_outflow_template.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_cooling_outflow_template", &resource.Sweeper{
		Name:         "netbox_cooling_outflow_template",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/cooling-outflow-templates/", "name")
		},
	})
}
