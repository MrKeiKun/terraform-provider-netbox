package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// NetBox 4.6 lets every choice of a choice set have a color.
func TestAccNetboxCustomFieldChoiceSet_colors(t *testing.T) {
	testAccSkipBelowVersion(t, "4.6.0")
	name := testAccGetTestName("choice-colors")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_custom_field_choice_set" "test" {
  name          = "%s"
  extra_choices = [["a", "A"], ["b", "B"]]
  choice_colors = {
    a = "red"
    b = "blue"
  }
}`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_custom_field_choice_set.test", "choice_colors.a", "red"),
					resource.TestCheckResourceAttr("netbox_custom_field_choice_set.test", "choice_colors.b", "blue"),
				),
			},
			{
				ResourceName:      "netbox_custom_field_choice_set.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
