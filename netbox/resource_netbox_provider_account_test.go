package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxProviderAccountDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_circuit_provider" "test" {
  name = "%[1]s"
}
`, testName)
}

func TestAccNetboxProviderAccount_basic(t *testing.T) {
	testName := testAccGetTestName("provideraccoun")
	deps := testAccNetboxProviderAccountDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_provider_account", providerAccountAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_provider_account" "test" {
  provider_id = netbox_circuit_provider.test.id
  account = "%[1]s"
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_provider_account.test", "account", testName),
					resource.TestCheckResourceAttr("netbox_provider_account.test", "name", testName),
				),
			},
			{
				ResourceName:      "netbox_provider_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxProviderAccountDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("provideraccoun_ds")
	deps := testAccNetboxProviderAccountDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_provider_account" "test" {
  provider_id = netbox_circuit_provider.test.id
  account = "%[1]s"
  name = "%[1]s"
}

data "netbox_provider_account" "test" {
  depends_on = [netbox_provider_account.test]
  account = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_provider_account.test", "id", "netbox_provider_account.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_provider_account.test", "account", "netbox_provider_account.test", "account"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_provider_account", &resource.Sweeper{
		Name:         "netbox_provider_account",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "circuits/provider-accounts/", "account")
		},
	})
}
