package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxCables() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxCablesRead),
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves all cables from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
			"cables": {
				Type:        schema.TypeList,
				Description: "The cables of the cables.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the cables.",
							Computed:    true,
						},
						"label": {
							Type:        schema.TypeString,
							Description: "The label of the cables.",
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "The type of the cables.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "The status of the cables.",
							Computed:    true,
						},
						"tenant_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the tenant associated with the cables.",
							Computed:    true,
						},
						"color": {
							Type:        schema.TypeString,
							Description: "The color of the cables as a hex string without the leading #.",
							Computed:    true,
						},
						"length": {
							Type:        schema.TypeFloat,
							Description: "The length of the cables.",
							Computed:    true,
						},
						"length_unit": {
							Type:        schema.TypeString,
							Description: "The length unit of the cables.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the cables.",
							Computed:    true,
						},
						"a_terminations": {
							Type:        schema.TypeList,
							Description: "The a terminations of the cables.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"object_type": {Type: schema.TypeString,
										Description: "The object type of the cables.", Computed: true},
									"object_id": {Type: schema.TypeInt,
										Description: "The ID of the object associated with the cables.", Computed: true},
								},
							},
						},
						"b_terminations": {
							Type:        schema.TypeList,
							Description: "The b terminations of the cables.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"object_type": {Type: schema.TypeString,
										Description: "The object type of the cables.", Computed: true},
									"object_id": {Type: schema.TypeInt,
										Description: "The ID of the object associated with the cables.", Computed: true},
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
