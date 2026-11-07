package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxCoolingOutflow() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxCoolingOutflowRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Requires NetBox v4.7.0 or newer.`,
		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"module_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"diameter": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"diameter_unit": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cooling_intake_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxCoolingOutflowRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addIntQuery(query, d, "device_id", "device_id")
	r, err := findOne[map[string]interface{}](m, "dcim/cooling-outflows/", query, "cooling outflow")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("device_id", jID(*r, "device"))
	d.Set("module_id", jID(*r, "module"))
	d.Set("name", jStr(*r, "name"))
	d.Set("label", jStr(*r, "label"))
	d.Set("type", jChoice(*r, "type"))
	d.Set("diameter", jFloatP(*r, "diameter"))
	d.Set("diameter_unit", jChoice(*r, "diameter_unit"))
	d.Set("cooling_intake_id", jID(*r, "cooling_intake"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
