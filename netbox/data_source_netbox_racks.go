package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxRacks() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxRacksRead),
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves all racks from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
			"racks": {
				Type:        schema.TypeList,
				Description: "The racks of the racks.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the racks.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the racks.",
							Computed:    true,
						},
						"site_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the site associated with the racks.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "The status of the racks.",
							Computed:    true,
						},
						"width": {
							Type:        schema.TypeInt,
							Description: "The width of the racks.",
							Computed:    true,
						},
						"u_height": {
							Type:        schema.TypeInt,
							Description: "The height of the racks in rack units.",
							Computed:    true,
						},
						tagsKey: tagsSchemaRead,
						"tenant_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the tenant associated with the racks.",
							Computed:    true,
						},
						"facility_id": {
							Type:        schema.TypeString,
							Description: "The ID of the facility associated with the racks.",
							Computed:    true,
						},
						"location_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the location associated with the racks.",
							Computed:    true,
						},
						"role_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the role associated with the racks.",
							Computed:    true,
						},
						"serial": {
							Type:        schema.TypeString,
							Description: "The serial number of the racks.",
							Computed:    true,
						},
						"asset_tag": {
							Type:        schema.TypeString,
							Description: "The asset tag of the racks.",
							Computed:    true,
						},
						"type_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the type associated with the racks.",
							Computed:    true,
						},
						"weight": {
							Type:        schema.TypeFloat,
							Description: "The weight of the racks.",
							Computed:    true,
						},
						"max_weight": {
							Type:        schema.TypeInt,
							Description: "The max weight of the racks.",
							Computed:    true,
						},
						"weight_unit": {
							Type:        schema.TypeString,
							Description: "The unit of the weight.",
							Computed:    true,
						},
						"desc_units": {
							Type:        schema.TypeBool,
							Description: "Whether desc units is enabled for the racks.",
							Computed:    true,
						},
						"outer_width": {
							Type:        schema.TypeInt,
							Description: "The outer width of the racks.",
							Computed:    true,
						},
						"outer_depth": {
							Type:        schema.TypeInt,
							Description: "The outer depth of the racks.",
							Computed:    true,
						},
						"outer_unit": {
							Type:        schema.TypeString,
							Description: "The outer unit of the racks.",
							Computed:    true,
						},
						"mounting_depth": {
							Type:        schema.TypeInt,
							Description: "The mounting depth of the racks.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the racks.",
							Computed:    true,
						},
						"comments": {
							Type:        schema.TypeString,
							Description: "Free-form comments about the racks.",
							Computed:    true,
						},
						"custom_fields": {
							Type:        schema.TypeMap,
							Description: "Custom field values of the racks, as a map of field name to value.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxRacksRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"asset_tag":      "asset_tag",
		"contact":        "contact",
		"contact_group":  "contact_group",
		"contact_role":   "contact_role",
		"desc_units":     "desc_units",
		"facility_id":    "facility_id",
		"id":             "id",
		"location_id":    "location_id",
		"max_weight":     "max_weight",
		"mounting_depth": "mounting_depth",
		"name":           "name",
		"outer_depth":    "outer_depth",
		"outer_unit":     "outer_unit",
		"outer_width":    "outer_width",
		"region_id":      "region_id",
		"role_id":        "role_id",
		"serial":         "serial",
		"site_id":        "site_id",
		"status":         "status",
		"tenant_id":      "tenant_id",
		"type_id":        "type",
		"u_height":       "u_height",
		"weight":         "weight",
		"weight_unit":    "weight_unit",
		"width":          "width",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "dcim/racks/", "racks", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["name"] = jStr(v, "name")
		mapping["site_id"] = jID(v, "site")
		mapping["status"] = jChoice(v, "status")
		mapping["width"] = jChoiceInt(v, "width")
		mapping["u_height"] = jInt(v, "u_height")
		mapping["tags"] = jTags(v)
		mapping["tenant_id"] = jID(v, "tenant")
		mapping["facility_id"] = jStr(v, "facility_id")
		mapping["location_id"] = jID(v, "location")
		mapping["role_id"] = jID(v, "role")
		mapping["serial"] = jStr(v, "serial")
		mapping["asset_tag"] = jStr(v, "asset_tag")
		mapping["type_id"] = jInt(v, "type")
		mapping["weight"] = jFloatP(v, "weight")
		mapping["max_weight"] = jIntP(v, "max_weight")
		mapping["desc_units"] = jBool(v, "desc_units")
		mapping["outer_width"] = jIntP(v, "outer_width")
		mapping["outer_depth"] = jIntP(v, "outer_depth")
		mapping["outer_unit"] = jChoice(v, "outer_unit")
		mapping["mounting_depth"] = jIntP(v, "mounting_depth")
		mapping["description"] = jStr(v, "description")
		mapping["comments"] = jStr(v, "comments")
		mapping["custom_fields"] = getCustomFields(v["custom_fields"])

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("racks", s)
}
