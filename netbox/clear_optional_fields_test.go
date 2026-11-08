package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Optional values that are removed from the configuration must be cleared in
// NetBox, not silently kept (upstream issue 857).
func TestAccNetboxPrefix_clearOptionalFields(t *testing.T) {
	name := testAccGetTestName("prefix_clear")
	deps := fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name = "%[1]s"
}
resource "netbox_vrf" "test" {
  name = "%[1]s"
}
resource "netbox_tag" "test" {
  name = "%[1]s"
}
`, name)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: deps + `
resource "netbox_prefix" "test" {
  prefix      = "10.77.0.0/24"
  status      = "active"
  description = "set"
  comments    = "set"
  tenant_id   = netbox_tenant.test.id
  vrf_id      = netbox_vrf.test.id
  tags        = [netbox_tag.test.name]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_prefix.test", "description", "set"),
					resource.TestCheckResourceAttrSet("netbox_prefix.test", "tenant_id"),
				),
			},
			{
				Config: deps + `
resource "netbox_prefix" "test" {
  prefix = "10.77.0.0/24"
  status = "active"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_prefix.test", "description", ""),
					resource.TestCheckResourceAttr("netbox_prefix.test", "comments", ""),
					testAccCheckAttrUnset("netbox_prefix.test", "tenant_id"),
					testAccCheckAttrUnset("netbox_prefix.test", "vrf_id"),
					resource.TestCheckResourceAttr("netbox_prefix.test", "tags.#", "0"),
				),
			},
		},
	})
}

func TestAccNetboxIPAddress_clearOptionalFields(t *testing.T) {
	name := testAccGetTestName("ip_clear")
	deps := fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name = "%[1]s"
}
resource "netbox_vrf" "test" {
  name = "%[1]s"
}
`, name)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: deps + `
resource "netbox_ip_address" "test" {
  ip_address  = "10.77.1.1/24"
  status      = "active"
  description = "set"
  dns_name    = "set.example.com"
  tenant_id   = netbox_tenant.test.id
  vrf_id      = netbox_vrf.test.id
}`,
				Check: resource.TestCheckResourceAttr("netbox_ip_address.test", "dns_name", "set.example.com"),
			},
			{
				Config: deps + `
resource "netbox_ip_address" "test" {
  ip_address = "10.77.1.1/24"
  status     = "active"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "description", ""),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "dns_name", ""),
					testAccCheckAttrUnset("netbox_ip_address.test", "tenant_id"),
					testAccCheckAttrUnset("netbox_ip_address.test", "vrf_id"),
				),
			},
		},
	})
}

func TestAccNetboxAvailableIPAddress_clearOptionalFields(t *testing.T) {
	name := testAccGetTestName("avip_clear")
	deps := fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name = "%[1]s"
}
resource "netbox_prefix" "test" {
  prefix = "10.77.2.0/24"
  status = "active"
}
`, name)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: deps + `
resource "netbox_available_ip_address" "test" {
  prefix_id   = netbox_prefix.test.id
  status      = "active"
  description = "set"
  dns_name    = "set.example.com"
  tenant_id   = netbox_tenant.test.id
}`,
				Check: resource.TestCheckResourceAttr("netbox_available_ip_address.test", "dns_name", "set.example.com"),
			},
			{
				Config: deps + `
resource "netbox_available_ip_address" "test" {
  prefix_id = netbox_prefix.test.id
  status    = "active"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_ip_address.test", "description", ""),
					resource.TestCheckResourceAttr("netbox_available_ip_address.test", "dns_name", ""),
					testAccCheckAttrUnset("netbox_available_ip_address.test", "tenant_id"),
				),
			},
		},
	})
}

// testAccCheckAttrUnset accepts both a missing attribute and a zero ID, which is
// how an unset relation is stored in the state.
func testAccCheckAttrUnset(name, attr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource %s not found", name)
		}
		if v := rs.Primary.Attributes[attr]; v != "" && v != "0" {
			return fmt.Errorf("%s.%s should be unset, got %q", name, attr, v)
		}
		return nil
	}
}
