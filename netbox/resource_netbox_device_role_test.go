package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxDeviceRole_basic(t *testing.T) {
	testSlug := "dvcrl_basic"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_device_role" "test" {
  name = "%s"
  slug = "%s"
  color_hex = "111111"
  description = "Some fancy device role"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_role.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_device_role.test", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_device_role.test", "color_hex", "111111"),
					resource.TestCheckResourceAttr("netbox_device_role.test", "description", "Some fancy device role"),
					// vm_role is not set, so it must settle at the default (false)
					// without a perpetual diff back to Netbox's server default.
					resource.TestCheckResourceAttr("netbox_device_role.test", "vm_role", "false"),
				),
			},
			{
				ResourceName:      "netbox_device_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetboxDeviceRole_vmRole is a regression test for the vm_role drift
// (#820): an explicit vm_role must round-trip in both directions without a
// perpetual plan. This only works when the client can send an explicit
// `false` (VMRole *bool), not a bool dropped by omitempty.
func TestAccNetboxDeviceRole_vmRole(t *testing.T) {
	testName := testAccGetTestName("dvcrl_vmrole")
	role := func(vmRole bool) string {
		return fmt.Sprintf(`
resource "netbox_device_role" "test" {
  name      = "%s"
  color_hex = "112233"
  vm_role   = %t
}`, testName, vmRole)
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: role(true),
				Check:  resource.TestCheckResourceAttr("netbox_device_role.test", "vm_role", "true"),
			},
			{
				Config: role(false),
				Check:  resource.TestCheckResourceAttr("netbox_device_role.test", "vm_role", "false"),
			},
		},
	})
}

func TestAccNetboxDeviceRole_defaultSlug(t *testing.T) {
	testSlug := "device_role_defSlug"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_device_role" "test" {
  name = "%s"
  color_hex = "111111"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_role.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_device_role.test", "slug", getSlug(testName)),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_device_role", &resource.Sweeper{
		Name:         "netbox_device_role",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "dcim/device-roles/", "name")
		},
	})
}
