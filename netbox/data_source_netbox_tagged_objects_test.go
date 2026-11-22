package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// NetBox returns the tagged objects in no fixed order, so reading them page by page could repeat some
// and miss others. With a page size of 1 every object is a page of its own.
func TestAccNetboxTaggedObjectsDataSource_pagedOneByOne(t *testing.T) {
	testName := testAccGetTestName("tagged_objects_paged")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "netbox" {
  page_size = 1
}

resource "netbox_tag" "test" {
  name = "%[1]s"
}

resource "netbox_site" "test" {
  count  = 6
  name   = "%[1]s-${count.index}"
  status = "active"
  tags   = [netbox_tag.test.name]
}

data "netbox_tagged_objects" "test" {
  filter {
    name  = "tag_id"
    value = netbox_tag.test.id
  }
  depends_on = [netbox_site.test]
}
`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_tagged_objects.test", "tagged_objects.#", "6"),
				),
			},
		},
	})
}
