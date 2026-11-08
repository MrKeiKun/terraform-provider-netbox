package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxCoolingIntake() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxCoolingIntakeRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Requires NetBox v4.7.0 or newer.`,
		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the cooling intake.",
				Optional:    true,
				Computed:    true,
			},
			"module_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module associated with the cooling intake.",
				Computed:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the cooling intake.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the cooling intake.",
				Computed:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the cooling intake.",
				Computed:    true,
			},
			"diameter": {
				Type:        schema.TypeFloat,
				Description: "The diameter of the cooling intake.",
				Computed:    true,
			},
			"diameter_unit": {
				Type:        schema.TypeString,
				Description: "The diameter unit of the cooling intake.",
				Computed:    true,
			},
			"max_flow": {
				Type:        schema.TypeFloat,
				Description: "The max flow of the cooling intake.",
				Computed:    true,
			},
			"max_flow_unit": {
				Type:        schema.TypeString,
				Description: "The max flow unit of the cooling intake.",
				Computed:    true,
			},
			"cooling_outflow_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the cooling outflow associated with the cooling intake.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cooling intake.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the cooling intake.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxCoolingIntakeRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addIntQuery(query, d, "device_id", "device_id")
	r, err := findOne[map[string]interface{}](m, "dcim/cooling-intakes/", query, "cooling intake")
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
	d.Set("max_flow", jFloatP(*r, "max_flow"))
	d.Set("max_flow_unit", jChoice(*r, "max_flow_unit"))
	d.Set("cooling_outflow_id", jID(*r, "cooling_outflow"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
