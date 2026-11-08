package netbox

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxToken_basic(t *testing.T) {
	if testAccNetboxVersionAtLeast("4.5.0") {
		t.Skipf("Skipping token test on NetBox %s: token creation requires API_TOKEN_PEPPERS which is not configured in the test environment", os.Getenv("NETBOX_VERSION"))
	}

	testSlug := "users"
	testName := testAccGetTestName(testSlug)
	testToken := testAccGetTestToken()
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_user" "test" {
  username = "%s"
  password = "Abcdefghijkl1"
}

resource "netbox_token" "test_basic" {
  user_id       = netbox_user.test.id
  key           = "%s"
  allowed_ips   = ["2.4.8.16/32"]
  write_enabled = false
  description   = "Netbox Test Basic Token"
  expires       = "2036-01-02T15:04:05.000Z"
}`, testName, testToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_token.test_basic", "key", testToken),
					resource.TestCheckResourceAttr("netbox_token.test_basic", "allowed_ips.#", "1"),
					resource.TestCheckResourceAttr("netbox_token.test_basic", "allowed_ips.0", "2.4.8.16/32"),
					resource.TestCheckResourceAttr("netbox_token.test_basic", "write_enabled", "false"),
					resource.TestCheckResourceAttr("netbox_token.test_basic", "description", "Netbox Test Basic Token"),
					resource.TestCheckResourceAttr("netbox_token.test_basic", "expires", "2036-01-02T15:04:05.000Z"),
				),
			},
			{
				ResourceName:            "netbox_token.test_basic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key"},
			},
		},
	})
}

func TestAccNetboxToken_withoutExpires(t *testing.T) {
	if testAccNetboxVersionAtLeast("4.5.0") {
		t.Skipf("Skipping token test on NetBox %s: token creation requires API_TOKEN_PEPPERS which is not configured in the test environment", os.Getenv("NETBOX_VERSION"))
	}
	testSlug := "users"
	testName := testAccGetTestName(testSlug)
	testToken := testAccGetTestToken()
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_user" "test" {
  username = "%s"
  password = "Abcdefghijkl1"
}

resource "netbox_token" "test_without_expires" {
  user_id       = netbox_user.test.id
  key           = "%s"
  allowed_ips   = ["2.4.8.16/32"]
  write_enabled = false
  description   = "Netbox Token Without Expires"
}`, testName, testToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_token.test_without_expires", "key", testToken),
					resource.TestCheckResourceAttr("netbox_token.test_without_expires", "allowed_ips.#", "1"),
					resource.TestCheckResourceAttr("netbox_token.test_without_expires", "allowed_ips.0", "2.4.8.16/32"),
					resource.TestCheckResourceAttr("netbox_token.test_without_expires", "write_enabled", "false"),
					resource.TestCheckResourceAttr("netbox_token.test_without_expires", "description", "Netbox Token Without Expires"),
					resource.TestCheckNoResourceAttr("netbox_token.test_without_expires", "expires"),
				),
			},
			{
				ResourceName:            "netbox_token.test_without_expires",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key"},
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_token", &resource.Sweeper{
		Name:         "netbox_token",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "users/users/", "username")
		},
	})
}

func TestAccNetboxToken_v2(t *testing.T) {
	if !testAccNetboxVersionAtLeast("4.5.0") {
		t.Skip("v2 tokens need NetBox 4.5 or later")
	}
	testName := testAccGetTestName("token_v2")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_user" "test" {
  username = "%s"
  password = "Abcdefghijkl1"
}

resource "netbox_token" "test" {
  user_id       = netbox_user.test.id
  write_enabled = false
  description   = "v2 token"
}

resource "netbox_token" "expiring" {
  user_id     = netbox_user.test.id
  description = "v2 token with expiry"
  expires     = "2036-01-02T15:04:05.000Z"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("netbox_token.test", "key"),
					// without expires the token must not expire (upstream issue 772)
					resource.TestCheckNoResourceAttr("netbox_token.test", "expires"),
					resource.TestCheckResourceAttr("netbox_token.expiring", "expires", "2036-01-02T15:04:05.000Z"),
					resource.TestCheckResourceAttr("netbox_token.test", "write_enabled", "false"),
					resource.TestCheckResourceAttr("netbox_token.test", "description", "v2 token"),
				),
			},
		},
	})
}
