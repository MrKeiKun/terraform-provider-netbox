package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxExportTemplateDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxExportTemplate_basic(t *testing.T) {
	testName := testAccGetTestName("exporttemplate")
	deps := testAccNetboxExportTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_export_template", exportTemplateAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_export_template" "test" {
  name = "%[1]s"
  object_types = ["dcim.site"]
  template_code = "{%% for s in queryset %%}{{ s.name }}\n{%% endfor %%}"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_export_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_export_template.test", "object_types.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_export_template" "test" {
  description = "updated"
  name = "%[1]s"
  object_types = ["dcim.site"]
  template_code = "{%% for s in queryset %%}{{ s.name }}\n{%% endfor %%}"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_export_template.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_export_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxExportTemplateDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("exporttemplate_ds")
	deps := testAccNetboxExportTemplateDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_export_template" "test" {
  name = "%[1]s"
  object_types = ["dcim.site"]
  template_code = "{%% for s in queryset %%}{{ s.name }}\n{%% endfor %%}"
}

data "netbox_export_template" "test" {
  depends_on = [netbox_export_template.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_export_template.test", "id", "netbox_export_template.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_export_template.test", "name", "netbox_export_template.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_export_template", &resource.Sweeper{
		Name:         "netbox_export_template",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "extras/export-templates/", "name")
		},
	})
}
