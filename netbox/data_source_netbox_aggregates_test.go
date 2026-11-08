package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxAggregatesDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("aggregates_ds")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_rir" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}

resource "netbox_tenant" "test" {
  name = "%[1]s"
}

resource "netbox_tag" "test" {
  name = "%[1]s"
}

resource "netbox_aggregate" "test" {
  prefix      = "10.77.0.0/16"
  description = "%[1]s_description"
  rir_id      = netbox_rir.test.id
  tenant_id   = netbox_tenant.test.id
  date_added  = "2024-01-02"
  tags        = [netbox_tag.test.name]
}

data "netbox_aggregates" "test" {
  depends_on = [netbox_aggregate.test]
  filter {
    name  = "prefix"
    value = "10.77.0.0/16"
  }
}

data "netbox_aggregates" "limited" {
  depends_on = [netbox_aggregate.test]
  limit      = 1
  filter {
    name  = "rir_id"
    value = netbox_rir.test.id
  }
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_aggregates.test", "aggregates.#", "1"),
					resource.TestCheckResourceAttrPair("data.netbox_aggregates.test", "aggregates.0.id", "netbox_aggregate.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_aggregates.test", "aggregates.0.prefix", "10.77.0.0/16"),
					resource.TestCheckResourceAttrPair("data.netbox_aggregates.test", "aggregates.0.rir_id", "netbox_rir.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_aggregates.test", "aggregates.0.tenant_id", "netbox_tenant.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_aggregates.test", "aggregates.0.date_added", "2024-01-02"),
					resource.TestCheckResourceAttr("data.netbox_aggregates.test", "aggregates.0.description", testName+"_description"),
					resource.TestCheckResourceAttr("data.netbox_aggregates.test", "aggregates.0.tags.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_aggregates.limited", "aggregates.#", "1"),
				),
			},
		},
	})
}

func TestAccNetboxAggregatesDataSource_empty(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
data "netbox_aggregates" "test" {
  filter {
    name  = "prefix"
    value = "10.78.0.0/16"
  }
}`,
				Check: resource.TestCheckResourceAttr("data.netbox_aggregates.test", "aggregates.#", "0"),
			},
		},
	})
}
