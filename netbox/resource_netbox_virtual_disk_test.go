package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccNetboxVirtualDisk_basic(t *testing.T) {
	testSlug := "virtual_disk"
	testName := testAccGetTestName(testSlug)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckVirtualDiskDestroy,
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
	return testAccCheckObjectsDestroyed(s, "netbox_virtual_disk", "virtualization/virtual-disks/")
}
