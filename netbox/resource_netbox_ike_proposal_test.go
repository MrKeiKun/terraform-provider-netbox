package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxIKEProposalDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxIKEProposal_basic(t *testing.T) {
	testName := testAccGetTestName("ikeproposal")
	deps := testAccNetboxIKEProposalDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_ike_proposal", iKEProposalAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ike_proposal" "test" {
  name = "%[1]s"
  authentication_method = "preshared-keys"
  encryption_algorithm = "aes-128-cbc"
  authentication_algorithm = "hmac-sha256"
  group = 14
  sa_lifetime = 3600
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ike_proposal.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_ike_proposal.test", "authentication_method", "preshared-keys"),
					resource.TestCheckResourceAttr("netbox_ike_proposal.test", "group", "14"),
					resource.TestCheckResourceAttr("netbox_ike_proposal.test", "sa_lifetime", "3600"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ike_proposal" "test" {
  description = "updated"
  name = "%[1]s"
  authentication_method = "preshared-keys"
  encryption_algorithm = "aes-128-cbc"
  authentication_algorithm = "hmac-sha256"
  group = 14
  sa_lifetime = 3600
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ike_proposal.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_ike_proposal.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxIKEProposalDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("ikeproposal_ds")
	deps := testAccNetboxIKEProposalDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ike_proposal" "test" {
  name = "%[1]s"
  authentication_method = "preshared-keys"
  encryption_algorithm = "aes-128-cbc"
  authentication_algorithm = "hmac-sha256"
  group = 14
  sa_lifetime = 3600
}

data "netbox_ike_proposal" "test" {
  depends_on = [netbox_ike_proposal.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_ike_proposal.test", "id", "netbox_ike_proposal.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_ike_proposal.test", "name", "netbox_ike_proposal.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_ike_proposal", &resource.Sweeper{
		Name:         "netbox_ike_proposal",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "vpn/ike-proposals/", "name")
		},
	})
}
