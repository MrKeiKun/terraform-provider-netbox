package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxIPSecProposalDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxIPSecProposal_basic(t *testing.T) {
	testName := testAccGetTestName("ipsecproposal")
	deps := testAccNetboxIPSecProposalDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_ipsec_proposal", iPSecProposalAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ipsec_proposal" "test" {
  name = "%[1]s"
  encryption_algorithm = "aes-128-cbc"
  authentication_algorithm = "hmac-sha256"
  sa_lifetime_seconds = 3600
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ipsec_proposal.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_ipsec_proposal.test", "encryption_algorithm", "aes-128-cbc"),
					resource.TestCheckResourceAttr("netbox_ipsec_proposal.test", "sa_lifetime_seconds", "3600"),
				),
			},
			{
				ResourceName:      "netbox_ipsec_proposal.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxIPSecProposalDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("ipsecproposal_ds")
	deps := testAccNetboxIPSecProposalDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_ipsec_proposal" "test" {
  name = "%[1]s"
  encryption_algorithm = "aes-128-cbc"
  authentication_algorithm = "hmac-sha256"
  sa_lifetime_seconds = 3600
}

data "netbox_ipsec_proposal" "test" {
  depends_on = [netbox_ipsec_proposal.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_ipsec_proposal.test", "id", "netbox_ipsec_proposal.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_ipsec_proposal.test", "name", "netbox_ipsec_proposal.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_ipsec_proposal", &resource.Sweeper{
		Name:         "netbox_ipsec_proposal",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "vpn/ipsec-proposals/", "name")
		},
	})
}
