package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxTagDataSource_basic(t *testing.T) {
	testSlug := "tag_ds_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name        = "%[1]s"
  description = "tag data source test"
}

data "netbox_tag" "test" {
  depends_on = [netbox_tag.test]
  name       = "%[1]s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_tag.test", "id", "netbox_tag.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_tag.test", "slug", "netbox_tag.test", "slug"),
					resource.TestCheckResourceAttr("data.netbox_tag.test", "description", "tag data source test"),
				),
			},
		},
	})
}

func TestAccNetboxTagDataSource_notFound(t *testing.T) {
	testName := testAccGetTestName("tag_ds_missing")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "netbox_tag" "test" {
  name = "%s"
}`, testName),
				ExpectError: regexp.MustCompile("no tag found matching filter"),
			},
		},
	})
}
