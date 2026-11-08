package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxCustomLinkDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxCustomLink_basic(t *testing.T) {
	testName := testAccGetTestName("customlink")
	deps := testAccNetboxCustomLinkDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_custom_link", customLinkAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_custom_link" "test" {
  name = "%[1]s"
  object_types = ["dcim.site"]
  link_text = "Open"
  link_url = "https://example.com/{{ object.name }}"
  weight = 50
  new_window = true
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_custom_link.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_custom_link.test", "link_text", "Open"),
					resource.TestCheckResourceAttr("netbox_custom_link.test", "weight", "50"),
					resource.TestCheckResourceAttr("netbox_custom_link.test", "new_window", "true"),
					resource.TestCheckResourceAttr("netbox_custom_link.test", "object_types.#", "1"),
				),
			},
			{
				ResourceName:      "netbox_custom_link.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCustomLinkDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("customlink_ds")
	deps := testAccNetboxCustomLinkDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_custom_link" "test" {
  name = "%[1]s"
  object_types = ["dcim.site"]
  link_text = "Open"
  link_url = "https://example.com/{{ object.name }}"
  weight = 50
  new_window = true
}

data "netbox_custom_link" "test" {
  depends_on = [netbox_custom_link.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_custom_link.test", "id", "netbox_custom_link.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_custom_link.test", "name", "netbox_custom_link.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_custom_link", &resource.Sweeper{
		Name:         "netbox_custom_link",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "extras/custom-links/", "name")
		},
	})
}
