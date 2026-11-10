package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxUser_basic(t *testing.T) {
	testSlug := "users"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_user" "test_basic" {
  username = "%s"
  password = "Abcdefghijkl1"
  active = true
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_user.test_basic", "username", testName),
					resource.TestCheckResourceAttr("netbox_user.test_basic", "active", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_user" "test_basic" {
  username = "%s"
  password = "Abcdefghijkl1"
	email = "foo@bar.com"
	first_name = "Hannah"
	last_name = "Acker"
  active = true
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_user.test_basic", "username", testName),
					resource.TestCheckResourceAttr("netbox_user.test_basic", "email", "foo@bar.com"),
					resource.TestCheckResourceAttr("netbox_user.test_basic", "first_name", "Hannah"),
					resource.TestCheckResourceAttr("netbox_user.test_basic", "last_name", "Acker"),
					resource.TestCheckResourceAttr("netbox_user.test_basic", "active", "true"),
				),
			},
			{
				ResourceName:      "netbox_user.test_basic",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccNetboxUser_group(t *testing.T) {
	testSlug := "users"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_group" "test_group" {
	  name = "%[1]s-group"
}

resource "netbox_user" "test_group" {
	  username = "%[1]s"
	  password = "Abcdefghijkl1"
	  active = true
	  group_ids = [netbox_group.test_group.id]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_user.test_group", "username", testName),
					resource.TestCheckResourceAttr("netbox_user.test_group", "active", "true"),
					resource.TestCheckResourceAttr("netbox_user.test_group", "group_ids.#", "1"),
				),
			},
			{
				ResourceName:      "netbox_user.test_group",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_user", &resource.Sweeper{
		Name:         "netbox_user",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "users/users/", "username")
		},
	})
}
