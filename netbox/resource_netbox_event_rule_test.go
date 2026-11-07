package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccNetboxEventRule_basic(t *testing.T) {
	testName := testAccGetTestName("evt_rule_basic")
	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetBoxEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_webhook" "test" {
  name        = "%[1]s"
  payload_url = "https://example.com/webhook"
}

resource "netbox_event_rule" "test" {
  name             = "%[1]s"
  description      = "foo description"
  content_types    = ["dcim.site"]
  action_type      = "webhook"
  action_object_id = netbox_webhook.test.id
  event_types      = ["object_created", "object_updated", "object_deleted", "job_started", "job_completed", "job_failed", "job_errored"]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_event_rule.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_event_rule.test", "content_types.#", "1"),
					resource.TestCheckResourceAttr("netbox_event_rule.test", "content_types.0", "dcim.site"),
					resource.TestCheckResourceAttr("netbox_event_rule.test", "action_type", "webhook"),
					resource.TestCheckResourceAttr("netbox_event_rule.test", "description", "foo description"),
					resource.TestCheckResourceAttr("netbox_event_rule.test", "event_types.#", "7"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_webhook" "test" {
  name        = "%[1]s"
  payload_url = "https://example.com/webhook"
}

resource "netbox_event_rule" "test" {
  name             = "%[1]s"
  content_types    = ["dcim.site", "virtualization.cluster"]
  action_type      = "webhook"
  action_object_id = netbox_webhook.test.id
  event_types      = ["object_created"]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_event_rule.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_event_rule.test", "content_types.#", "2"),
					resource.TestCheckResourceAttr("netbox_event_rule.test", "content_types.0", "dcim.site"),
					resource.TestCheckResourceAttr("netbox_event_rule.test", "content_types.1", "virtualization.cluster"),
					resource.TestCheckResourceAttr("netbox_event_rule.test", "action_type", "webhook"),
					resource.TestCheckResourceAttr("netbox_event_rule.test", "event_types.#", "1"),
				),
			},
			{
				ResourceName:      "netbox_event_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckNetBoxEventRuleDestroy(s *terraform.State) error {
	return testAccCheckObjectsDestroyed(s, "netbox_event_rule", "extras/event-rules/")
}

func init() {
	resource.AddTestSweepers("netbox_event_rule", &resource.Sweeper{
		Name:         "netbox_event_rule",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "extras/event-rules/", "name")
		},
	})
}
