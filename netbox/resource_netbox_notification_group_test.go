package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxNotificationGroupDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_user" "test" {
  username = "%[1]s"
  password = "Abcdefghijkl1"
}
`, testName)
}

func TestAccNetboxNotificationGroup_basic(t *testing.T) {
	testName := testAccGetTestName("notificationgr")
	deps := testAccNetboxNotificationGroupDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_notification_group", "extras/notification-groups/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_notification_group" "test" {
  name = "%[1]s"
  user_ids = [netbox_user.test.id]
  description = "test"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_notification_group.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_notification_group.test", "description", "test"),
					resource.TestCheckResourceAttr("netbox_notification_group.test", "user_ids.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_notification_group" "test" {
  name = "%[1]s"
  user_ids = [netbox_user.test.id]
  description = "test-updated"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_notification_group.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_notification_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxNotificationGroupDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("notificationgr_ds")
	deps := testAccNetboxNotificationGroupDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_notification_group" "test" {
  name = "%[1]s"
  user_ids = [netbox_user.test.id]
  description = "test"
}

data "netbox_notification_group" "test" {
  depends_on = [netbox_notification_group.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_notification_group.test", "id", "netbox_notification_group.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_notification_group.test", "name", "netbox_notification_group.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_notification_group", &resource.Sweeper{
		Name:         "netbox_notification_group",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "extras/notification-groups/", "name")
		},
	})
}
