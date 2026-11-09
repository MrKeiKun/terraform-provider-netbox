package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxIKEPolicyDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_ike_proposal" "test" {
  name                 = "%[1]s"
  authentication_method = "preshared-keys"
  encryption_algorithm = "aes-128-cbc"
  group                = 14
}
`, testName)
}

func TestAccNetboxIKEPolicy_basic(t *testing.T) {
	testName := testAccGetTestName("ikepolicy")
	deps := testAccNetboxIKEPolicyDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_ike_policy", iKEPolicyAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ike_policy" "test" {
  name = "%[1]s"
  version = 2
  proposal_ids = [netbox_ike_proposal.test.id]
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ike_policy.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_ike_policy.test", "version", "2"),
					resource.TestCheckResourceAttr("netbox_ike_policy.test", "proposal_ids.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ike_policy" "test" {
  description = "updated"
  name = "%[1]s"
  version = 2
  proposal_ids = [netbox_ike_proposal.test.id]
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ike_policy.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_ike_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxIKEPolicyDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("ikepolicy_ds")
	deps := testAccNetboxIKEPolicyDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ike_policy" "test" {
  name = "%[1]s"
  version = 2
  proposal_ids = [netbox_ike_proposal.test.id]
}

data "netbox_ike_policy" "test" {
  depends_on = [netbox_ike_policy.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_ike_policy.test", "id", "netbox_ike_policy.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_ike_policy.test", "name", "netbox_ike_policy.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_ike_policy", &resource.Sweeper{
		Name:         "netbox_ike_policy",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "vpn/ike-policies/", "name")
		},
	})
}
