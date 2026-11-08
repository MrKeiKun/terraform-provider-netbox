package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxDeviceModuleBays() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxDeviceModuleBaysRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves all device module bays from NetBox. Narrow the result with filter blocks and cap it with limit.`,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeSet,
				Description: "Filters that select which items are returned. Each filter is a name/value pair.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the field to filter by.",
							Required:    true,
						},
						"value": {
							Type:        schema.TypeString,
							Description: "The value of the field to filter by.",
							Required:    true,
						},
					},
				},
			},
			"limit": {
				Type:             schema.TypeInt,
				Description:      "The maximum number of items to return. Defaults to 0, which returns all items.",
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
			},
			"module_bays": {
				Type:        schema.TypeList,
				Description: "The module bays of the device module bays.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the device module bays.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the device module bays.",
							Computed:    true,
						},
						"device_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the device associated with the device module bays.",
							Computed:    true,
						},
						"label": {
							Type:        schema.TypeString,
							Description: "The label of the device module bays.",
							Computed:    true,
						},
						"position": {
							Type:        schema.TypeString,
							Description: "The position of the device module bays.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the device module bays.",
							Computed:    true,
						},
						"installed_module_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the installed module associated with the device module bays.",
							Computed:    true,
						},
						"tags": tagsSchemaRead,
					},
				},
			},
		},
	}
}

func dataSourceNetboxDeviceModuleBaysRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"name":      "name",
		"name__ic":  "name__ic",
		"device_id": "device_id",
		"device":    "device",
		"tag":       "tag",
		"position":  "position",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "dcim/module-bays/", "module bays", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["name"] = jStr(v, "name")
		mapping["device_id"] = jID(v, "device")
		mapping["label"] = jStr(v, "label")
		mapping["position"] = jStr(v, "position")
		mapping["description"] = jStr(v, "description")
		mapping["installed_module_id"] = jID(v, "installed_module")
		mapping["tags"] = jTags(v)

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("module_bays", s)
}
