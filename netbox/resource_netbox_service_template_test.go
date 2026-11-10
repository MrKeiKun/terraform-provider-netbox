package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxServiceTemplateDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxServiceTemplate_basic(t *testing.T) {
	testName := testAccGetTestName("servicetemplat")
	deps := testAccNetboxServiceTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_service_template", serviceTemplateAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_service_template" "test" {
  name = "%[1]s"
  protocol = "tcp"
  ports = [80, 443]
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_service_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_service_template.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("netbox_service_template.test", "ports.#", "2"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_service_template" "test" {
  description = "updated"
  name = "%[1]s"
  protocol = "tcp"
  ports = [80, 443]
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_service_template.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_service_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxServiceTemplateDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("servicetemplat_ds")
	deps := testAccNetboxServiceTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_service_template" "test" {
  name = "%[1]s"
  protocol = "tcp"
  ports = [80, 443]
}

data "netbox_service_template" "test" {
  depends_on = [netbox_service_template.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_service_template.test", "id", "netbox_service_template.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_service_template.test", "name", "netbox_service_template.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_service_template", &resource.Sweeper{
		Name:         "netbox_service_template",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "ipam/service-templates/", "name")
		},
	})
}
