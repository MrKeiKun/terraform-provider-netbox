package netbox

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccNetboxAggregatesSetUp() string {
	return `
resource "netbox_rir" "test" {
  name = "test_rir"
}

resource "netbox_tenant" "test" {
  name = "test_tenant"
}

resource "netbox_aggregate" "test_1" {
  prefix = "192.168.0.0/16"
  rir_id = netbox_rir.test.id
}

resource "netbox_aggregate" "test_2" {
  prefix      = "10.0.0.0/8"
  description = "Test aggregate"
  rir_id      = netbox_rir.test.id
  tenant_id   = netbox_tenant.test.id
}

resource "netbox_aggregate" "test_3" {
  prefix = "172.16.0.0/12"
  rir_id = netbox_rir.test.id
}`
}

func testAccNetboxAggregatesByPrefix() string {
	return `
data "netbox_aggregates" "test" {
  filter {
	name  = "prefix"
	value = "192.168.0.0/16"
  }
}`
}

func testAccNetboxAggregatesByDescription() string {
	return `
data "netbox_aggregates" "test" {
  filter {
	name  = "description"
	value = "Test aggregate"
  }
}`
}

func TestAccNetboxAggregatesDataSource_basic(t *testing.T) {
	setUp := testAccNetboxAggregatesSetUp()
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: setUp,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_aggregate.test_1", "prefix", "192.168.0.0/16"),
					resource.TestCheckResourceAttr("netbox_aggregate.test_2", "prefix", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("netbox_aggregate.test_3", "prefix", "172.16.0.0/12"),
				),
			},
			{
				Config: setUp + testAccNetboxAggregatesByPrefix(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_aggregates.test", "aggregates.#", "1"),
					resource.TestCheckResourceAttrPair("data.netbox_aggregates.test", "aggregates.0.prefix", "netbox_aggregate.test_1", "prefix"),
				),
			},
			{
				Config: setUp + testAccNetboxAggregatesByDescription(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_aggregates.test", "aggregates.#", "1"),
					resource.TestCheckResourceAttrPair("data.netbox_aggregates.test", "aggregates.0.description", "netbox_aggregate.test_2", "description"),
				),
			},
		},
	})
}
