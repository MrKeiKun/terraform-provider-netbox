package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxConfigContextProfileDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxConfigContextProfile_basic(t *testing.T) {
	testName := testAccGetTestName("configcontextp")
	deps := testAccNetboxConfigContextProfileDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_config_context_profile", configContextProfileAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_config_context_profile" "test" {
  name = "%[1]s"
  description = "test"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_config_context_profile.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_config_context_profile.test", "description", "test"),
				),
			},
			{
				ResourceName:      "netbox_config_context_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxConfigContextProfileDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("configcontextp_ds")
	deps := testAccNetboxConfigContextProfileDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_config_context_profile" "test" {
  name = "%[1]s"
  description = "test"
}

data "netbox_config_context_profile" "test" {
  depends_on = [netbox_config_context_profile.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_config_context_profile.test", "id", "netbox_config_context_profile.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_config_context_profile.test", "name", "netbox_config_context_profile.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_config_context_profile", &resource.Sweeper{
		Name:         "netbox_config_context_profile",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "extras/config-context-profiles/", "name")
		},
	})
}
