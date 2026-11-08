package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxTenantDataSource_tags(t *testing.T) {
	testName := testAccGetTestName("tenant_ds_tags")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%[1]s"
}

resource "netbox_tenant" "test" {
  name = "%[1]s"
  tags = [netbox_tag.test.name]
}

data "netbox_tenant" "by_name" {
  name = netbox_tenant.test.name
}

data "netbox_tenants" "all" {
  depends_on = [netbox_tenant.test]
  filter {
    name  = "name"
    value = "%[1]s"
  }
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_tenant.by_name", "tags.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_tenant.by_name", "tags.0", testName),
					resource.TestCheckResourceAttr("data.netbox_tenants.all", "tenants.0.tags.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_tenants.all", "tenants.0.tags.0", testName),
				),
			},
		},
	})
}

func TestAccNetboxVlansDataSource_siteFilter(t *testing.T) {
	testName := testAccGetTestName("vlans_ds_site")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = "%[1]s"
  slug   = "%[1]s"
  status = "active"
}

resource "netbox_vlan" "at_site" {
  name    = "%[1]s_at"
  vid     = 2201
  site_id = netbox_site.test.id
}

resource "netbox_vlan" "elsewhere" {
  name = "%[1]s_else"
  vid  = 2202
}

data "netbox_vlans" "in_site" {
  depends_on = [netbox_vlan.at_site, netbox_vlan.elsewhere]
  filter {
    name  = "site"
    value = netbox_site.test.slug
  }
}

data "netbox_vlans" "not_in_site" {
  depends_on = [netbox_vlan.at_site, netbox_vlan.elsewhere]
  filter {
    name  = "site__n"
    value = netbox_site.test.slug
  }
  filter {
    name  = "vid"
    value = "2202"
  }
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_vlans.in_site", "vlans.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_vlans.in_site", "vlans.0.vid", "2201"),
					resource.TestCheckResourceAttr("data.netbox_vlans.not_in_site", "vlans.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_vlans.not_in_site", "vlans.0.vid", "2202"),
				),
			},
		},
	})
}

func TestAccNetboxIPAddressesDataSource_assignedObject(t *testing.T) {
	testName := testAccGetTestName("ips_ds_assigned")
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxDeviceInterfaceFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_device_interface" "test" {
  name      = "%[1]s"
  device_id = netbox_device.test.id
  type      = "1000base-t"
  mgmtonly  = true
}

resource "netbox_ip_address" "test" {
  ip_address   = "10.92.1.1/24"
  status       = "active"
  interface_id = netbox_device_interface.test.id
  object_type  = "dcim.interface"
}

data "netbox_ip_addresses" "test" {
  depends_on = [netbox_ip_address.test]
  filter {
    name  = "ip_address"
    value = "10.92.1.1/24"
  }
}

data "netbox_device_interfaces" "test" {
  depends_on = [netbox_device_interface.test]
  filter {
    name  = "device_id"
    value = netbox_device.test.id
  }
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_ip_addresses.test", "ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_ip_addresses.test", "ip_addresses.0.assigned_object_type", "dcim.interface"),
					resource.TestCheckResourceAttrPair("data.netbox_ip_addresses.test", "ip_addresses.0.assigned_object_id", "netbox_device_interface.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_device_interfaces.test", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_device_interfaces.test", "interfaces.0.mgmt_only", "true"),
				),
			},
		},
	})
}
