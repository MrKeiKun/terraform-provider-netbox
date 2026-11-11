package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxIPSecProfileDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_ike_policy" "test" {
  name    = "%[1]s"
  version = 2
}

resource "netbox_ipsec_policy" "test" {
  name = "%[1]s"
}
`, testName)
}

func TestAccNetboxIPSecProfile_basic(t *testing.T) {
	testName := testAccGetTestName("ipsecprofile")
	deps := testAccNetboxIPSecProfileDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_ipsec_profile", "vpn/ipsec-profiles/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ipsec_profile" "test" {
  name = "%[1]s"
  mode = "esp"
  ike_policy_id = netbox_ike_policy.test.id
  ipsec_policy_id = netbox_ipsec_policy.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ipsec_profile.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_ipsec_profile.test", "mode", "esp"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ipsec_profile" "test" {
  description = "updated"
  name = "%[1]s"
  mode = "esp"
  ike_policy_id = netbox_ike_policy.test.id
  ipsec_policy_id = netbox_ipsec_policy.test.id
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ipsec_profile.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_ipsec_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxIPSecProfileDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("ipsecprofile_ds")
	deps := testAccNetboxIPSecProfileDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ipsec_profile" "test" {
  name = "%[1]s"
  mode = "esp"
  ike_policy_id = netbox_ike_policy.test.id
  ipsec_policy_id = netbox_ipsec_policy.test.id
}

data "netbox_ipsec_profile" "test" {
  depends_on = [netbox_ipsec_profile.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_ipsec_profile.test", "id", "netbox_ipsec_profile.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_ipsec_profile.test", "name", "netbox_ipsec_profile.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_ipsec_profile", &resource.Sweeper{
		Name:         "netbox_ipsec_profile",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "vpn/ipsec-profiles/", "name")
		},
	})
}
