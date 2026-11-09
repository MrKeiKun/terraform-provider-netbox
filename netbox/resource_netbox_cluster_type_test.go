package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxClusterType_basic(t *testing.T) {
	testSlug := "clstr_type_data_basic"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_cluster_type" "test" {
  name = "%s"
  slug = "%s"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cluster_type.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cluster_type.test", "slug", randomSlug),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_cluster_type" "test" {
  description = "updated"
  name = "%s"
  slug = "%s"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cluster_type.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_cluster_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxClusterType_defaultSlug(t *testing.T) {
	testSlug := "clstr_type_data_default_slug"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_cluster_type" "test" {
  name = "%s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cluster_type.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cluster_type.test", "slug", testName),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_cluster_type", &resource.Sweeper{
		Name:         "netbox_cluster_type",
		Dependencies: []string{"netbox_cluster"},
		F: func(region string) error {
			return sweepByName(region, "virtualization/cluster-types/", "name")
		},
	})
}
