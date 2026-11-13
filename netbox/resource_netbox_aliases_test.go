package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// netbox_virtual_machine_interface is the same resource as the deprecated netbox_interface.
func TestAccNetboxVirtualMachineInterface_basic(t *testing.T) {
	name := testAccGetTestName("vm-interface")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}

resource "netbox_virtual_machine" "test" {
  name    = "%[1]s"
  site_id = netbox_site.test.id
}

resource "netbox_virtual_machine_interface" "new" {
  name               = "eth0"
  virtual_machine_id = netbox_virtual_machine.test.id
  mtu                = 1500
}

resource "netbox_interface" "old" {
  name               = "eth1"
  virtual_machine_id = netbox_virtual_machine.test.id
}
`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine_interface.new", "name", "eth0"),
					resource.TestCheckResourceAttr("netbox_virtual_machine_interface.new", "mtu", "1500"),
					resource.TestCheckResourceAttrPair("netbox_virtual_machine_interface.new", "virtual_machine_id", "netbox_virtual_machine.test", "id"),
					resource.TestCheckResourceAttr("netbox_interface.old", "name", "eth1"),
				),
			},
			{
				ResourceName:      "netbox_virtual_machine_interface.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCircuitProviderAccount_basic(t *testing.T) {
	name := testAccGetTestName("provider-account")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_circuit_provider" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}

resource "netbox_circuit_provider_account" "test" {
  provider_id = netbox_circuit_provider.test.id
  account     = "%[1]s"
}
`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit_provider_account.test", "account", name),
					resource.TestCheckResourceAttrPair("netbox_circuit_provider_account.test", "provider_id", "netbox_circuit_provider.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_circuit_provider_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// A resource that did not have tags before takes them now.
func TestAccNetboxManufacturer_tags(t *testing.T) {
	name := testAccGetTestName("manufacturer-tags")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}

resource "netbox_manufacturer" "test" {
  name = "%[1]s"
  slug = "%[1]s"
  tags = [netbox_tag.test.name]
}
`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "tags_all.#", "1"),
				),
			},
			{
				ResourceName:      "netbox_manufacturer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// A moved block turns a netbox_interface into a netbox_virtual_machine_interface without replacing it.
func TestAccNetboxVirtualMachineInterface_moved(t *testing.T) {
	name := testAccGetTestName("vm-interface-moved")
	base := fmt.Sprintf(`
resource "netbox_site" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}

resource "netbox_virtual_machine" "test" {
  name    = "%[1]s"
  site_id = netbox_site.test.id
}
`, name)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: base + `
resource "netbox_interface" "eth0" {
  name               = "eth0"
  virtual_machine_id = netbox_virtual_machine.test.id
}
`,
			},
			{
				Config: base + `
moved {
  from = netbox_interface.eth0
  to   = netbox_virtual_machine_interface.eth0
}

resource "netbox_virtual_machine_interface" "eth0" {
  name               = "eth0"
  virtual_machine_id = netbox_virtual_machine.test.id
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("netbox_virtual_machine_interface.eth0", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.TestCheckResourceAttr("netbox_virtual_machine_interface.eth0", "name", "eth0"),
			},
		},
	})
}
