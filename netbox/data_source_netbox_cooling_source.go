package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxCoolingSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxCoolingSourceRead),
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Requires NetBox v4.7.0 or newer.`,
		Schema: map[string]*schema.Schema{
			"site_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the site associated with the cooling source.",
				Computed:    true,
			},
			"location_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the location associated with the cooling source.",
				Computed:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the cooling source.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the cooling source.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the cooling source.",
				Computed:    true,
			},
			"fluid_type": {
				Type:        schema.TypeString,
				Description: "The fluid type of the cooling source.",
				Computed:    true,
			},
			"cooling_capacity": {
				Type:        schema.TypeFloat,
				Description: "The cooling capacity of the cooling source.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cooling source.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the cooling source.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the cooling source.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxCoolingSourceRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "dcim/cooling-sources/", query, "cooling source")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("site_id", jID(*r, "site"))
	d.Set("location_id", jID(*r, "location"))
	d.Set("name", jStr(*r, "name"))
	d.Set("type", jChoice(*r, "type"))
	d.Set("status", jChoice(*r, "status"))
	d.Set("fluid_type", jChoice(*r, "fluid_type"))
	d.Set("cooling_capacity", jFloatP(*r, "cooling_capacity"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
