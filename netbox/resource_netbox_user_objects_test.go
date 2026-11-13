package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// userObjectsConfig has a user and a site that the bookmark, notification and subscription point at.
func userObjectsConfig(name string) string {
	return fmt.Sprintf(`
resource "netbox_user" "test" {
  username = "%[1]s"
  password = "Abcdefghijkl1"
  active   = true
}

resource "netbox_site" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}
`, name)
}

func TestAccNetboxBookmark_basic(t *testing.T) {
	name := testAccGetTestName("bookmark")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: userObjectsConfig(name) + `
resource "netbox_bookmark" "test" {
  object_type = "dcim.site"
  object_id   = netbox_site.test.id
  user_id     = netbox_user.test.id
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_bookmark.test", "object_type", "dcim.site"),
					resource.TestCheckResourceAttrPair("netbox_bookmark.test", "object_id", "netbox_site.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_bookmark.test", "user_id", "netbox_user.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_bookmark.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxSubscription_basic(t *testing.T) {
	name := testAccGetTestName("subscription")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: userObjectsConfig(name) + `
resource "netbox_subscription" "test" {
  object_type = "dcim.site"
  object_id   = netbox_site.test.id
  user_id     = netbox_user.test.id
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_subscription.test", "object_type", "dcim.site"),
					resource.TestCheckResourceAttrPair("netbox_subscription.test", "user_id", "netbox_user.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_subscription.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxNotification_basic(t *testing.T) {
	name := testAccGetTestName("notification")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: userObjectsConfig(name) + `
resource "netbox_notification" "test" {
  object_type = "dcim.site"
  object_id   = netbox_site.test.id
  user_id     = netbox_user.test.id
  event_type  = "object_updated"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_notification.test", "event_type", "object_updated"),
					resource.TestCheckResourceAttrPair("netbox_notification.test", "user_id", "netbox_user.test", "id"),
				),
			},
			{
				Config: userObjectsConfig(name) + `
resource "netbox_notification" "test" {
  object_type = "dcim.site"
  object_id   = netbox_site.test.id
  user_id     = netbox_user.test.id
  event_type  = "object_created"
}`,
				Check: resource.TestCheckResourceAttr("netbox_notification.test", "event_type", "object_created"),
			},
			{
				ResourceName:      "netbox_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxTableConfig_basic(t *testing.T) {
	name := testAccGetTestName("table-config")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_table_config" "test" {
  name        = "%s"
  object_type = "dcim.site"
  table       = "SiteTable"
  columns     = ["name", "status"]
}`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_table_config.test", "name", name),
					resource.TestCheckResourceAttr("netbox_table_config.test", "columns.#", "2"),
					resource.TestCheckResourceAttr("netbox_table_config.test", "columns.0", "name"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_table_config" "test" {
  name        = "%s"
  object_type = "dcim.site"
  table       = "SiteTable"
  columns     = ["name", "status", "region"]
  ordering    = ["name"]
  description = "Sites with region"
  shared      = true
}`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_table_config.test", "columns.#", "3"),
					resource.TestCheckResourceAttr("netbox_table_config.test", "ordering.0", "name"),
					resource.TestCheckResourceAttr("netbox_table_config.test", "description", "Sites with region"),
					resource.TestCheckResourceAttr("netbox_table_config.test", "shared", "true"),
				),
			},
			{
				ResourceName:      "netbox_table_config.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
