package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxCables() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxCablesRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):`,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"limit": {
				Type:             schema.TypeInt,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
			},
			"cables": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"label": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tenant_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"color": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"length": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"length_unit": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"a_terminations": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"object_type": {Type: schema.TypeString, Computed: true},
									"object_id":   {Type: schema.TypeInt, Computed: true},
								},
							},
						},
						"b_terminations": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"object_type": {Type: schema.TypeString, Computed: true},
									"object_id":   {Type: schema.TypeInt, Computed: true},
								},
							},
						},
						"tags": tagsSchemaRead,
					},
				},
			},
		},
	}
}

func dataSourceNetboxCablesRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"label":        "label",
		"type":         "type",
		"status":       "status",
		"color":        "color",
		"tenant_id":    "tenant_id",
		"device_id":    "device_id",
		"device":       "device",
		"rack_id":      "rack_id",
		"site_id":      "site_id",
		"site":         "site",
		"location_id":  "location_id",
		"tag":          "tag",
		"unterminated": "unterminated",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "dcim/cables/", "cables", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["label"] = jStr(v, "label")
		mapping["type"] = jChoice(v, "type")
		mapping["status"] = jChoice(v, "status")
		mapping["tenant_id"] = jID(v, "tenant")
		mapping["color"] = jStr(v, "color")
		mapping["length"] = jFloatP(v, "length")
		mapping["length_unit"] = jChoice(v, "length_unit")
		mapping["description"] = jStr(v, "description")
		mapping["a_terminations"] = jObjectsList(v, "a_terminations", "object_type", "object_id")
		mapping["b_terminations"] = jObjectsList(v, "b_terminations", "object_type", "object_id")
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
	return d.Set("cables", s)
}
