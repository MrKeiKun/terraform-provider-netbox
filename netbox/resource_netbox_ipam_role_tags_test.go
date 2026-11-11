package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxRole_tags(t *testing.T) {
	testName := testAccGetTestName("role_tags")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%[1]s"
}

resource "netbox_ipam_role" "test" {
  name = "%[1]s"
  slug = "%[1]s"
  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ipam_role.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_ipam_role.test", "tags.0", testName),
				),
			},
			{
				ResourceName:      "netbox_ipam_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%[1]s"
}

resource "netbox_ipam_role" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}`, testName),
				Check: resource.TestCheckResourceAttr("netbox_ipam_role.test", "tags.#", "0"),
			},
		},
	})
}
