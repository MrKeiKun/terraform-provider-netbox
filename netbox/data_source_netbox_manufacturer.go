package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxManufacturer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetboxManufacturerRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Fetches the list of manufacturers from netbox. Can optionally be filtered by name, slug or tag.

From the [official documentation](https://netboxlabs.com/docs/netbox/models/dcim/manufacturer/):

> A manufacturer represents the "make" of a device; e.g. Cisco or Dell. Each device type must be assigned to a manufacturer. (Inventory items and platforms may also be associated with manufacturers.)`,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      `The name of the field to filter by. Supported values are "name", "slug" and "tag".`,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"name", "slug", "tag"}, false)),
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: `The value to filter by. For "name" and "slug" this is a string, for "tag" this is the name of the tag.`,
						},
					},
				},
			},
			"limit": {
				Type:             schema.TypeInt,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
				Description:      `The maximum number of items to return. Will return all items if not set`,
			},
			"manufacturers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slug": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxManufacturerRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"name": "name",
		"tag":  "tag",
		"slug": "slug",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "dcim/manufacturers/", "manufacturers", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["description"] = jStr(v, "description")
		mapping["slug"] = jStr(v, "slug")
		mapping["name"] = jStr(v, "name")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("manufacturers", s)
}
