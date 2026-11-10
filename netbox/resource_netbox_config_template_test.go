package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxConfigTemplate_basic(t *testing.T) {
	testName := testAccGetTestName("config_template")
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_config_template" "test" {
	name = "%[1]s"
	description = "%[1]s description"
	template_code = "hostname {{ device }}"
	environment_params = jsonencode({"trim_blocks" = true})
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_config_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_config_template.test", "description", fmt.Sprintf("%s description", testName)),
					resource.TestCheckResourceAttr("netbox_config_template.test", "template_code", "hostname {{ device }}"),
					resource.TestCheckResourceAttr("netbox_config_template.test", "environment_params", "{\"trim_blocks\":true}"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_config_template" "test" {
	name = "%[1]s"
	description = "%[1]s description"
	template_code = "hostname {{ device }}"
	environment_params = jsonencode({"lstrip_blocks" = true})
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_config_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_config_template.test", "description", fmt.Sprintf("%s description", testName)),
					resource.TestCheckResourceAttr("netbox_config_template.test", "template_code", "hostname {{ device }}"),
					resource.TestCheckResourceAttr("netbox_config_template.test", "environment_params", "{\"lstrip_blocks\":true}"),
				),
			},
			{
				ResourceName:      "netbox_config_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxConfigTemplate_tags(t *testing.T) {
	testName := testAccGetTestName("config_template_tags")
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxVrfTagDependencies(testName) + fmt.Sprintf(`
resource "netbox_config_template" "test_tags" {
  name = "%[1]s"
  template_code = "hostname test"
  tags = [netbox_tag.test_a.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_config_template.test_tags", "name", testName),
					resource.TestCheckResourceAttr("netbox_config_template.test_tags", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_config_template.test_tags", "tags.0", testName+"a"),
				),
			},
			{
				Config: testAccNetboxVrfTagDependencies(testName) + fmt.Sprintf(`
resource "netbox_config_template" "test_tags" {
  name = "%[1]s"
  template_code = "hostname test"
  tags = [netbox_tag.test_a.name, netbox_tag.test_b.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_config_template.test_tags", "tags.#", "2"),
					resource.TestCheckResourceAttr("netbox_config_template.test_tags", "tags.0", testName+"a"),
					resource.TestCheckResourceAttr("netbox_config_template.test_tags", "tags.1", testName+"b"),
				),
			},
			{
				Config: testAccNetboxVrfTagDependencies(testName) + fmt.Sprintf(`
resource "netbox_config_template" "test_tags" {
  name = "%s"
  template_code = "hostname test"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_config_template.test_tags", "tags.#", "0"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_config_template", &resource.Sweeper{
		Name:         "netbox_config_template",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "extras/config-templates/", "name")
		},
	})
}
