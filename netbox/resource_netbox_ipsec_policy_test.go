package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxIPSecPolicyDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_ipsec_proposal" "test" {
  name                 = "%[1]s"
  encryption_algorithm = "aes-128-cbc"
}
`, testName)
}

func TestAccNetboxIPSecPolicy_basic(t *testing.T) {
	testName := testAccGetTestName("ipsecpolicy")
	deps := testAccNetboxIPSecPolicyDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_ipsec_policy", "vpn/ipsec-policies/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ipsec_policy" "test" {
  name = "%[1]s"
  pfs_group = 14
  proposal_ids = [netbox_ipsec_proposal.test.id]
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ipsec_policy.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_ipsec_policy.test", "pfs_group", "14"),
					resource.TestCheckResourceAttr("netbox_ipsec_policy.test", "proposal_ids.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ipsec_policy" "test" {
  description = "updated"
  name = "%[1]s"
  pfs_group = 14
  proposal_ids = [netbox_ipsec_proposal.test.id]
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ipsec_policy.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_ipsec_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxIPSecPolicyDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("ipsecpolicy_ds")
	deps := testAccNetboxIPSecPolicyDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ipsec_policy" "test" {
  name = "%[1]s"
  pfs_group = 14
  proposal_ids = [netbox_ipsec_proposal.test.id]
}

data "netbox_ipsec_policy" "test" {
  depends_on = [netbox_ipsec_policy.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_ipsec_policy.test", "id", "netbox_ipsec_policy.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_ipsec_policy.test", "name", "netbox_ipsec_policy.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_ipsec_policy", &resource.Sweeper{
		Name:         "netbox_ipsec_policy",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "vpn/ipsec-policies/", "name")
		},
	})
}
