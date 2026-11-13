package netbox

import (
	"fmt"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccNetboxTenantTagDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_tag" "test_a" {
  name = "%[1]sa"
}

resource "netbox_tag" "test_b" {
  name = "%[1]sb"
}
`, testName)
}

func TestAccNetboxTenant_basic(t *testing.T) {
	testSlug := "tenant_basic"
	testName := testAccGetTestName(testSlug)
	testDescription := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name = "%s"
  slug = "%s"
  description = "%s"
}`, testName, randomSlug, testDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tenant.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_tenant.test", "slug", randomSlug),
					resource.TestCheckResourceAttr("netbox_tenant.test", "description", testDescription),
				),
			},
			{
				ResourceName:      "netbox_tenant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxTenant_defaultSlug(t *testing.T) {
	testSlug := "tenant_defSlug"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name = "%s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tenant.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_tenant.test", "slug", fwresource.Slug(testName)),
				),
			},
		},
	})
}

func TestAccNetboxTenant_tags(t *testing.T) {
	testSlug := "tenant_tags"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxTenantTagDependencies(testName) + fmt.Sprintf(`
resource "netbox_tenant" "test_tags" {
  name = "%[1]s"
  tags = [netbox_tag.test_a.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tenant.test_tags", "name", testName),
					resource.TestCheckResourceAttr("netbox_tenant.test_tags", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_tenant.test_tags", "tags.0", testName+"a"),
				),
			},
			{
				Config: testAccNetboxTenantTagDependencies(testName) + fmt.Sprintf(`
resource "netbox_tenant" "test_tags" {
  name = "%[1]s"
  tags = [netbox_tag.test_a.name, netbox_tag.test_b.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tenant.test_tags", "tags.#", "2"),
					resource.TestCheckResourceAttr("netbox_tenant.test_tags", "tags.0", testName+"a"),
					resource.TestCheckResourceAttr("netbox_tenant.test_tags", "tags.1", testName+"b"),
				),
			},
			{
				Config: testAccNetboxTenantTagDependencies(testName) + fmt.Sprintf(`
resource "netbox_tenant" "test_tags" {
  name = "%s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tenant.test_tags", "tags.#", "0"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_tenant", &resource.Sweeper{
		Name:         "netbox_tenant",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "tenancy/tenants/", "name")
		},
	})
}
