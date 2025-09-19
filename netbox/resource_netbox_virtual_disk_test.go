package netbox

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/fbreckle/go-netbox/netbox/client/virtualization"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccNetboxVirtualDisk_basic(t *testing.T) {
	testSlug := "virtual_disk"
	testName := testAccGetTestName(testSlug)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "tag_a" {
	name = "[%[1]s_a]"
	color_hex = "123456"
}
resource "netbox_site" "test" {
	name = "%[1]s"
	status = "active"
}
resource "netbox_virtual_machine" "test" {
	name = "%[1]s"
	site_id = netbox_site.test.id
}
resource "netbox_virtual_disk" "test" {
	name = "%[1]s"
	description = "description"
	size_mb = 30
	virtual_machine_id = netbox_virtual_machine.test.id
	tags = [netbox_tag.tag_a.name]
}
				`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.#", "1"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.name", testName),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.description", "description"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.size_mb", "30"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.tags.0", "["+testName+"_a]"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "tag_a" {
	name = "[%[1]s_a]"
	color_hex = "123456"
}
resource "netbox_site" "test" {
	name = "%[1]s"
	status = "active"
}
resource "netbox_virtual_machine" "test" {
	name = "%[1]s"
	site_id = netbox_site.test.id
}
resource "netbox_virtual_disk" "test" {
	name = "%[1]s_updated"
	description = "description updated"
	size_mb = 60
	virtual_machine_id = netbox_virtual_machine.test.id
	tags = [netbox_tag.tag_a.name]
}
				`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.#", "1"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.name", testName+"_updated"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.description", "description updated"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.size_mb", "60"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.tags.0", "["+testName+"_a]"),
				),
			},
			{
				ResourceName:      "netbox_virtual_disk.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVirtualDiskDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*providerState)

	// loop through the resources in state, verifying each virtual machine
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "netbox_virtual_disk" {
			continue
		}

		stateID, _ := strconv.ParseInt(rs.Primary.ID, 10, 64)
		params := virtualization.NewVirtualizationVirtualDisksReadParams().WithID(stateID)
		_, err := conn.Virtualization.VirtualizationVirtualDisksRead(params, nil)

		if err == nil {
			return fmt.Errorf("virtual disk (%s) still exists", rs.Primary.ID)
		}

		if errresp, ok := err.(*virtualization.VirtualizationVirtualDisksReadDefault); ok {
			errorcode := errresp.Code()
			if errorcode == 404 {
				return nil
			}
		}
		return err
	}
	return nil
}

func TestAccNetboxVirtualMachine_virtualDisks(t *testing.T) {
	testSlug := "vm_virtual_disks"
	testName := testAccGetTestName(testSlug)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site" "test" {
	name = "%[1]s"
	status = "active"
}
resource "netbox_virtual_machine" "test" {
	name = "%[1]s"
	site_id = netbox_site.test.id
}
resource "netbox_virtual_disk" "disk1" {
	name = "%[1]s_disk1"
	description = "First disk"
	size_mb = 100
	virtual_machine_id = netbox_virtual_machine.test.id
}
resource "netbox_virtual_disk" "disk2" {
	name = "%[1]s_disk2"
	description = "Second disk"
	size_mb = 200
	virtual_machine_id = netbox_virtual_machine.test.id
}
				`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.#", "2"),
					// Check first disk
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.name", testName+"_disk1"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.description", "First disk"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.0.size_mb", "100"),
					// Check second disk
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.1.name", testName+"_disk2"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.1.description", "Second disk"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "virtual_disks.1.size_mb", "200"),
				),
			},
		},
	})
}
