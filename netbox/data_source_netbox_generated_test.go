package netbox

import (
	"fmt"
	"strings"
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwprovider"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The data sources that are made from the API schema find objects by their values and by filters.
func TestAccNetboxGeneratedDataSources(t *testing.T) {
	name := testAccGetTestName("generated-ds")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}

resource "netbox_manufacturer" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}

resource "netbox_device_type" "test" {
  model           = "%[1]s"
  slug            = "%[1]s"
  manufacturer_id = netbox_manufacturer.test.id
}

resource "netbox_device_role" "test" {
  name      = "%[1]s"
  slug      = "%[1]s"
  color_hex = "123456"
}

resource "netbox_device" "test" {
  name           = "%[1]s"
  device_type_id = netbox_device_type.test.id
  role_id        = netbox_device_role.test.id
  site_id        = netbox_site.test.id
}

resource "netbox_device_interface" "test" {
  name      = "eth0"
  device_id = netbox_device.test.id
  type      = "1000base-t"
}

data "netbox_manufacturer" "test" {
  name = netbox_manufacturer.test.name
}

data "netbox_sites" "test" {
  filter {
    name  = "name"
    value = netbox_site.test.name
  }
}

data "netbox_device_interface" "test" {
  name      = netbox_device_interface.test.name
  device_id = netbox_device.test.id
}

data "netbox_device_interfaces" "test" {
  depends_on = [netbox_device_interface.test]

  filter {
    name  = "device_id"
    value = netbox_device.test.id
  }
}

data "netbox_users" "all" {
  limit = 1
}
`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_manufacturer.test", "id", "netbox_manufacturer.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_sites.test", "sites.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_sites.test", "sites.0.name", name),
					resource.TestCheckResourceAttrPair("data.netbox_device_interface.test", "id", "netbox_device_interface.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_device_interface.test", "type", "1000base-t"),
					resource.TestCheckResourceAttr("data.netbox_device_interfaces.test", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_users.all", "users.#", "1"),
				),
			},
		},
	})
}

// An object without a name is found by its number.
func TestAccNetboxGeneratedDataSources_byID(t *testing.T) {
	name := testAccGetTestName("generated-ds-id")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site" "test" {
  name = "%[1]s"
  slug = "%[1]s"
}

resource "netbox_journal_entry" "test" {
  assigned_object_type = "dcim.site"
  assigned_object_id   = netbox_site.test.id
  comments             = "%[1]s"
}

resource "netbox_mac_address" "test" {
  mac_address = "00:1A:2B:3C:4D:5E"
}

data "netbox_journal_entry" "test" {
  id = netbox_journal_entry.test.id
}

data "netbox_mac_address" "test" {
  id = netbox_mac_address.test.id
}
`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_journal_entry.test", "id", "netbox_journal_entry.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_journal_entry.test", "comments", name),
					resource.TestCheckResourceAttrPair("data.netbox_mac_address.test", "id", "netbox_mac_address.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_mac_address.test", "mac_address", "00:1A:2B:3C:4D:5E"),
				),
			},
		},
	})
}

// Every data source that lists objects reads without a filter, on every NetBox version that has the
// object type. This also checks that what the API returns fits the attributes.
func TestAccNetboxGeneratedDataSources_listAll(t *testing.T) {
	_, dataSources := fwprovider.Inventory()
	var config strings.Builder
	count := 0
	for _, d := range dataSources {
		if d.Plural == nil {
			continue
		}
		if d.SinceVersion != "" && !testAccNetboxVersionAtLeast(d.SinceVersion) {
			continue
		}
		if hasRequired(d) {
			continue // it cannot be read without something to look for
		}
		count++
		fmt.Fprintf(&config, "data %q %q {\n", d.TypeName, "all")
		if !d.Plural.NoFilter {
			config.WriteString("  limit = 5\n")
		}
		config.WriteString("}\n")
	}
	if count < 50 {
		t.Fatalf("only %d data sources that list objects", count)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config.String(),
				Check:  resource.TestCheckResourceAttrSet("data.netbox_sites.all", "id"),
			},
		},
	})
}

func hasRequired(d *fwresource.DataSourceSpec) bool {
	for _, list := range [][]fwresource.Attr{d.Attrs, d.ExtraFilters, d.FilterBlock} {
		for _, a := range list {
			if a.Required {
				return true
			}
		}
	}
	return false
}
