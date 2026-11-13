package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccNetboxAvailableASN_basic(t *testing.T) {
	name := testAccGetTestName("available-asn")
	config := func(description string) string {
		return fmt.Sprintf(`
resource "netbox_rir" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}

resource "netbox_asn_range" "test" {
  name   = "%[1]s"
  slug   = "%[1]s"
  rir_id = netbox_rir.test.id
  start  = 64512
  end    = 64520
}

resource "netbox_available_asn" "first" {
  asn_range_id = netbox_asn_range.test.id
  description  = "%[2]s"
}

resource "netbox_available_asn" "second" {
  asn_range_id = netbox_asn_range.test.id
  description  = "second"
}
`, name, description)
	}
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config("first"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("netbox_available_asn.first", "asn"),
					resource.TestCheckResourceAttrSet("netbox_available_asn.second", "asn"),
					resource.TestCheckResourceAttrPair("netbox_available_asn.first", "rir_id", "netbox_rir.test", "id"),
					func(s *terraform.State) error {
						a := s.RootModule().Resources["netbox_available_asn.first"].Primary.Attributes["asn"]
						b := s.RootModule().Resources["netbox_available_asn.second"].Primary.Attributes["asn"]
						if a == b {
							return fmt.Errorf("both resources got the AS number %s", a)
						}
						return nil
					},
				),
			},
			{
				// changing the description keeps the AS number
				Config: config("changed"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_asn.first", "description", "changed"),
					resource.TestCheckResourceAttrSet("netbox_available_asn.first", "asn"),
				),
			},
			{
				ResourceName: "netbox_available_asn.first",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return availableImportID(s, "netbox_asn_range.test", "netbox_available_asn.first")
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"asn_range_id"},
			},
		},
	})
}

// availableImportID builds the "<parent id> <object id>" import ID of an available resource.
func availableImportID(s *terraform.State, parent, object string) (string, error) {
	p, ok := s.RootModule().Resources[parent]
	if !ok {
		return "", fmt.Errorf("%s not found", parent)
	}
	o, ok := s.RootModule().Resources[object]
	if !ok {
		return "", fmt.Errorf("%s not found", object)
	}
	return p.Primary.ID + " " + o.Primary.ID, nil
}
