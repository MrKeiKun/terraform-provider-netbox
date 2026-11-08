package netbox

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxDevicePowerPorts() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxDevicePowerPortRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves all device power ports from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
			"name_regex": {
				Type:         schema.TypeString,
				Description:  "The name regex of the device power ports.",
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"limit": {
				Type:             schema.TypeInt,
				Description:      "The maximum number of items to return. Defaults to 0, which returns all items.",
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
			},
			"power_ports": {
				Type:        schema.TypeList,
				Description: "The power ports of the device power ports.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the device power ports.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the device power ports.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the device power ports.",
							Computed:    true,
						},
						"tag_ids": {
							Type:        schema.TypeList,
							Description: "The IDs of the tag associated with the device power ports.",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"device_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the device associated with the device power ports.",
							Computed:    true,
						},
						"module_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the module associated with the device power ports.",
							Optional:    true,
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "The type of the device power ports.",
							Optional:    true,
							Computed:    true,
						},
						"maximum_draw": {
							Type:        schema.TypeInt,
							Description: "The maximum draw of the device power ports.",
							Optional:    true,
							Computed:    true,
						},
						"allocated_draw": {
							Type:        schema.TypeInt,
							Description: "The allocated draw of the device power ports.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxDevicePowerPortRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"name":      "name",
		"tag":       "tag",
		"device_id": "device_id",
	}

	// Fetch everything and apply the regular expression before the limit
	var keep func(map[string]interface{}) bool
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		keep = func(o map[string]interface{}) bool { return r.MatchString(jStr(o, "name")) }
	}

	results, err := listFilteredWith[map[string]interface{}](d, m, "dcim/power-ports/", "power ports", allowed, nil, keep)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["description"] = jStr(v, "description")
		mapping["name"] = jStr(v, "name")
		mapping["tag_ids"] = jTagIDs(v)
		mapping["device_id"] = jID(v, "device")
		mapping["module_id"] = jID(v, "module")
		mapping["type"] = jChoice(v, "type")
		mapping["maximum_draw"] = jIntP(v, "maximum_draw")
		mapping["allocated_draw"] = jIntP(v, "allocated_draw")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("power_ports", s)
}
