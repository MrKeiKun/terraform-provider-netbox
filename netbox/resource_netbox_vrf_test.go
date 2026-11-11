package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccNetboxVrfTagDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_tag" "test_a" {
  name = "%[1]sa"
}

resource "netbox_tag" "test_b" {
  name = "%[1]sb"
}
`, testName)
}

func testAccNetboxVrfTenantDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_tenant" "test_tenant_a" {
  name = "%[1]sa"
}
resource "netbox_tenant" "test_tenant_b" {
  name = "%[1]sb"
}
`, testName)
}

func TestAccNetboxVrf_basic(t *testing.T) {
	testSlug := "vrf_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_vrf" "test" {
  name        = "%s"
  description = "my-description"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vrf.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_vrf.test", "description", "my-description"),
				),
			},
			{
				ResourceName:      "netbox_vrf.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxVrf_tags(t *testing.T) {
	testSlug := "vrf_tag"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxVrfTagDependencies(testName) + fmt.Sprintf(`
resource "netbox_vrf" "test_tags" {
  name = "%[1]s"
  tags = [netbox_tag.test_a.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vrf.test_tags", "name", testName),
					resource.TestCheckResourceAttr("netbox_vrf.test_tags", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_vrf.test_tags", "tags.0", testName+"a"),
				),
			},
			{
				Config: testAccNetboxVrfTagDependencies(testName) + fmt.Sprintf(`
resource "netbox_vrf" "test_tags" {
  name = "%[1]s"
  tags = [netbox_tag.test_a.name, netbox_tag.test_b.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vrf.test_tags", "tags.#", "2"),
					resource.TestCheckResourceAttr("netbox_vrf.test_tags", "tags.0", testName+"a"),
					resource.TestCheckResourceAttr("netbox_vrf.test_tags", "tags.1", testName+"b"),
				),
			},
			{
				Config: testAccNetboxVrfTagDependencies(testName) + fmt.Sprintf(`
resource "netbox_vrf" "test_tags" {
  name = "%s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vrf.test_tags", "tags.#", "0"),
				),
			},
		},
	})
}

func TestAccNetboxVrf_tenant(t *testing.T) {
	testSlug := "vrf_tenant"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxVrfTenantDependencies(testName) + fmt.Sprintf(`
resource "netbox_vrf" "test_tenant" {
  name = "%[1]s"
  tenant_id = netbox_tenant.test_tenant_a.id
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vrf.test_tenant", "name", testName),
					resource.TestCheckResourceAttrPair("netbox_vrf.test_tenant", "tenant_id", "netbox_tenant.test_tenant_a", "id"),
				),
			},
			{
				Config: testAccNetboxVrfTenantDependencies(testName) + fmt.Sprintf(`
resource "netbox_vrf" "test_tenant" {
  name = "%[1]s"
  tenant_id = netbox_tenant.test_tenant_b.id
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vrf.test_tenant", "name", testName),
					resource.TestCheckResourceAttrPair("netbox_vrf.test_tenant", "tenant_id", "netbox_tenant.test_tenant_b", "id"),
				),
			},
		},
	})
}

func TestAccNetboxVrf_rd(t *testing.T) {
	testSlug := "vrf_rd"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_vrf" "test_rd" {
	name        = "%s"
	rd          = "123:456"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vrf.test_rd", "name", testName),
					resource.TestCheckResourceAttr("netbox_vrf.test_rd", "rd", "123:456"),
				),
			},
			{
				ResourceName:      "netbox_vrf.test_rd",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxVrf_enforceUnique(t *testing.T) {
	testSlug := "vrf_enforce_unique"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_vrf" "test_enforce_unique" {
	name        = "%s-true"
	enforce_unique = true
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vrf.test_enforce_unique", "name", testName+"-true"),
					resource.TestCheckResourceAttr("netbox_vrf.test_enforce_unique", "enforce_unique", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_vrf" "test_enforce_unique_false" {
	name        = "%s-false"
	enforce_unique = false
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vrf.test_enforce_unique_false", "name", testName+"-false"),
					resource.TestCheckResourceAttr("netbox_vrf.test_enforce_unique_false", "enforce_unique", "false"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_vrf", &resource.Sweeper{
		Name:         "netbox_vrf",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "ipam/vrfs/", "name")
		},
	})
}
