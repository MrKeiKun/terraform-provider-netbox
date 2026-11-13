package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxVlanTranslationRuleDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_vlan_translation_policy" "test" {
  name = "%[1]s"
}
`, testName)
}

func TestAccNetboxVlanTranslationRule_basic(t *testing.T) {
	testName := testAccGetTestName("vlantranslatio")
	deps := testAccNetboxVlanTranslationRuleDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_vlan_translation_rule", "ipam/vlan-translation-rules/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_vlan_translation_rule" "test" {
  policy_id = netbox_vlan_translation_policy.test.id
  local_vid = 100
  remote_vid = 200
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan_translation_rule.test", "local_vid", "100"),
					resource.TestCheckResourceAttr("netbox_vlan_translation_rule.test", "remote_vid", "200"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_vlan_translation_rule" "test" {
  description = "updated"
  policy_id = netbox_vlan_translation_policy.test.id
  local_vid = 100
  remote_vid = 200
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan_translation_rule.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_vlan_translation_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
