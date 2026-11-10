package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxCoolingIntakeTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxCoolingIntakeTemplateRead),
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Requires NetBox v4.7.0 or newer.`,
		Schema: map[string]*schema.Schema{
			"device_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device type associated with the cooling intake template.",
				Optional:    true,
				Computed:    true,
			},
			"module_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module type associated with the cooling intake template.",
				Computed:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the cooling intake template.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the cooling intake template.",
				Computed:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the cooling intake template.",
				Computed:    true,
			},
			"diameter": {
				Type:        schema.TypeFloat,
				Description: "The diameter of the cooling intake template.",
				Computed:    true,
			},
			"diameter_unit": {
				Type:        schema.TypeString,
				Description: "The diameter unit of the cooling intake template.",
				Computed:    true,
			},
			"max_flow": {
				Type:        schema.TypeFloat,
				Description: "The max flow of the cooling intake template.",
				Computed:    true,
			},
			"max_flow_unit": {
				Type:        schema.TypeString,
				Description: "The max flow unit of the cooling intake template.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cooling intake template.",
				Computed:    true,
			},
		},
	}
}

func dataSourceNetboxCoolingIntakeTemplateRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addIntQuery(query, d, "device_type_id", "device_type_id")
	r, err := findOne[map[string]interface{}](m, "dcim/cooling-intake-templates/", query, "cooling intake template")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("device_type_id", jID(*r, "device_type"))
	d.Set("module_type_id", jID(*r, "module_type"))
	d.Set("name", jStr(*r, "name"))
	d.Set("label", jStr(*r, "label"))
	d.Set("type", jChoice(*r, "type"))
	d.Set("diameter", jFloatP(*r, "diameter"))
	d.Set("diameter_unit", jChoice(*r, "diameter_unit"))
	d.Set("max_flow", jFloatP(*r, "max_flow"))
	d.Set("max_flow_unit", jChoice(*r, "max_flow_unit"))
	d.Set("description", jStr(*r, "description"))
	return nil
}
