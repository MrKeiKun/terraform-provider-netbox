package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxLocations() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxLocationsRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves all locations from NetBox. Narrow the result with filter blocks and cap it with limit.`,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A list of filter to apply to the API query when requesting locations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the field to filter on. Supported fields are: .",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The value to pass to the specified filter.",
						},
					},
				},
			},
			"tags": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "A list of tags to filter on.",
			},
			"limit": {
				Type:             schema.TypeInt,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
				Description:      "The limit of objects to return from the API lookup.",
			},
			"locations": {
				Type:        schema.TypeList,
				Description: "The locations of the locations.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "The numeric ID of the locations.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the locations.",
							Optional:    true,
						},
						"slug": {
							Type:        schema.TypeString,
							Description: "The URL-friendly slug of the locations.",
							Optional:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the locations.",
							Computed:    true,
						},
						"facility": {
							Type:        schema.TypeString,
							Description: "The facility of the locations.",
							Computed:    true,
						},
						"tenant_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the tenant associated with the locations.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "The status of the locations.",
							Computed:    true,
						},
						"site_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the site associated with the locations.",
							Computed:    true,
						},
						"parent_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the parent associated with the locations.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxLocationsRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"name":      "name",
		"slug":      "slug",
		"site":      "site",
		"site_id":   "site_id",
		"parent_id": "parent_id",
		"tenant":    "tenant",
		"tenant_id": "tenant_id",
		"status":    "status",
	}

	extra := url.Values{}
	if tags, ok := d.GetOk("tags"); ok {
		for _, tag := range tags.(*schema.Set).List() {
			extra.Add("tag", tag.(string))
		}
	}

	results, err := listFiltered[map[string]interface{}](d, m, "dcim/locations/", "locations", allowed, extra)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = strconv.FormatInt(jInt(v, "id"), 10)
		mapping["name"] = jStr(v, "name")
		mapping["slug"] = jStr(v, "slug")
		mapping["site_id"] = jID(v, "site")
		mapping["description"] = jStr(v, "description")
		mapping["facility"] = jStr(v, "facility")
		mapping["parent_id"] = jID(v, "parent")
		mapping["status"] = jChoice(v, "status")
		mapping["tenant_id"] = jID(v, "tenant")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("locations", s)
}
