package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccNetboxVlanTranslationPolicyDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxVlanTranslationPolicy_basic(t *testing.T) {
	testName := testAccGetTestName("vlantranslatio")
	deps := testAccNetboxVlanTranslationPolicyDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_vlan_translation_policy", vlanTranslationPolicyAPIPath)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_vlan_translation_policy" "test" {
  name = "%[1]s"
  description = "test"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan_translation_policy.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_vlan_translation_policy.test", "description", "test"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_vlan_translation_policy" "test" {
  name = "%[1]s"
  description = "test-updated"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_vlan_translation_policy.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_vlan_translation_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxVlanTranslationPolicyDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("vlantranslatio_ds")
	deps := testAccNetboxVlanTranslationPolicyDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_vlan_translation_policy" "test" {
  name = "%[1]s"
  description = "test"
}

data "netbox_vlan_translation_policy" "test" {
  depends_on = [netbox_vlan_translation_policy.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_vlan_translation_policy.test", "id", "netbox_vlan_translation_policy.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_vlan_translation_policy.test", "name", "netbox_vlan_translation_policy.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_vlan_translation_policy", &resource.Sweeper{
		Name:         "netbox_vlan_translation_policy",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "ipam/vlan-translation-policies/", "name")
		},
	})
}
