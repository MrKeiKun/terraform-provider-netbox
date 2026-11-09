package netbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxAsn_basic(t *testing.T) {
	testSlug := "asn_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%[1]sa"
}

resource "netbox_rir" "test" {
  name = "%[1]s"
}

resource "netbox_asn" "test" {
  asn    = 1337
  rir_id = netbox_rir.test.id

  description = "test"
  comments = "test"

  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_asn.test", "asn", "1337"),
					resource.TestCheckResourceAttr("netbox_asn.test", "description", "test"),
					resource.TestCheckResourceAttr("netbox_asn.test", "comments", "test"),
					resource.TestCheckResourceAttr("netbox_asn.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_asn.test", "tags.0", testName+"a"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%[1]sa"
}

resource "netbox_rir" "test" {
  name = "%[1]s"
}

resource "netbox_asn" "test" {
  asn    = 1337
  rir_id = netbox_rir.test.id

  description = "test-updated"
  comments = "test"

  tags = [netbox_tag.test.name]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_asn.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_asn.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//func TestAccNetboxAsn_customFields(t *testing.T) {
//	testSlug := "asn_detail"
//	testName := testAccGetTestName(testSlug)
//	testField := strings.ReplaceAll(testAccGetTestName(testSlug), "-", "_")
//	resource.Test(t, resource.TestCase{
//		PreCheck:  func() { testAccPreCheck(t) },
//		Providers: testAccProviders,
//		Steps: []resource.TestStep{
//			{
//				Config: fmt.Sprintf(`
//resource "netbox_custom_field" "test" {
//	name          = "%[1]s"
//	type          = "text"
//	content_types = ["ipam.asn"]
//}
//resource "netbox_asn" "test" {
//  name          = "%[2]s"
//  status        = "active"
//  latitude      = "12.123456"
//  longitude     = "-13.123456"
//  timezone      = "Africa/Johannesburg"
//  custom_fields = {"${netbox_custom_field.test.name}" = "81"}
//}`, testField, testName),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr("netbox_asn.test", "custom_fields."+testField, "81"),
//					resource.TestCheckResourceAttr("netbox_asn.test", "timezone", "Africa/Johannesburg"),
//					resource.TestCheckResourceAttr("netbox_asn.test", "latitude", "12.123456"),
//					resource.TestCheckResourceAttr("netbox_asn.test", "longitude", "-13.123456"),
//				),
//			},
//		},
//	})
//}

func init() {
	resource.AddTestSweepers("netbox_asn", &resource.Sweeper{
		Name:         "netbox_asn",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "ipam/asns/", "asn")
		},
	})
}
