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
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the cooling outflow.",
				Optional:    true,
				Computed:    true,
			},
			"module_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module associated with the cooling outflow.",
				Computed:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the cooling outflow.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the cooling outflow.",
				Computed:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the cooling outflow.",
				Computed:    true,
			},
			"diameter": {
				Type:        schema.TypeFloat,
				Description: "The diameter of the cooling outflow.",
				Computed:    true,
			},
			"diameter_unit": {
				Type:        schema.TypeString,
				Description: "The diameter unit of the cooling outflow.",
				Computed:    true,
			},
			"cooling_intake_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the cooling intake associated with the cooling outflow.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cooling outflow.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the cooling outflow.",
				Computed:    true,
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
