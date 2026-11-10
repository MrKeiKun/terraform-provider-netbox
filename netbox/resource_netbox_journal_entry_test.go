package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxJournalEntryDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = "%[1]s"
  status = "active"
}
`, testName)
}

func TestAccNetboxJournalEntry_basic(t *testing.T) {
	testName := testAccGetTestName("journalentry")
	deps := testAccNetboxJournalEntryDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_journal_entry", journalEntryAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_journal_entry" "test" {
  assigned_object_type = "dcim.site"
  assigned_object_id = netbox_site.test.id
  kind = "info"
  comments = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_journal_entry.test", "kind", "info"),
					resource.TestCheckResourceAttr("netbox_journal_entry.test", "comments", testName),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_journal_entry" "test" {
  assigned_object_type = "dcim.site"
  assigned_object_id = netbox_site.test.id
  kind = "info"
  comments = "updated"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_journal_entry.test", "comments", "updated"),
				),
			},
			{
				ResourceName:      "netbox_journal_entry.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
