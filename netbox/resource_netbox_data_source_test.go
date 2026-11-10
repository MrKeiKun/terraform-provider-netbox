package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxDataSourceDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("datasource")
	deps := testAccNetboxDataSourceDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_data_source", dataSourceAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_data_source" "test" {
  name = "%[1]s"
  type = "git"
  source_url = "https://example.com/%[1]s.git"
  enabled = false
  sync_interval = 1440
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_data_source.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_data_source.test", "type", "git"),
					resource.TestCheckResourceAttr("netbox_data_source.test", "sync_interval", "1440"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_data_source" "test" {
  description = "updated"
  name = "%[1]s"
  type = "git"
  source_url = "https://example.com/%[1]s.git"
  enabled = false
  sync_interval = 1440
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_data_source.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_data_source.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxDataSourceDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("datasource_ds")
	deps := testAccNetboxDataSourceDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_data_source" "test" {
  name = "%[1]s"
  type = "git"
  source_url = "https://example.com/%[1]s.git"
  enabled = false
  sync_interval = 1440
}

data "netbox_data_source" "test" {
  depends_on = [netbox_data_source.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_data_source.test", "id", "netbox_data_source.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_data_source.test", "name", "netbox_data_source.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_data_source", &resource.Sweeper{
		Name:         "netbox_data_source",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "core/data-sources/", "name")
		},
	})
}
