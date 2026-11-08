package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxAggregates() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxAggregatesRead,
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves all aggregates from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
			"aggregates": {
				Type:        schema.TypeList,
				Description: "The aggregates of the aggregates.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the aggregates.",
							Computed:    true,
						},
						"prefix": {
							Type:        schema.TypeString,
							Description: "The prefix in CIDR notation.",
							Computed:    true,
						},
						"rir_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the rir associated with the aggregates.",
							Computed:    true,
						},
						"tenant_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the tenant associated with the aggregates.",
							Computed:    true,
						},
						"date_added": {
							Type:        schema.TypeString,
							Description: "The date added of the aggregates.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the aggregates.",
							Computed:    true,
						},
						"tags": tagsSchemaRead,
					},
				},
			},
		},
	}
}

func dataSourceNetboxAggregatesRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"prefix":     "prefix",
		"family":     "family",
		"rir_id":     "rir_id",
		"tenant_id":  "tenant_id",
		"tag":        "tag",
		"q":          "q",
		"date_added": "date_added",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "ipam/aggregates/", "aggregates", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["prefix"] = jStr(v, "prefix")
		mapping["rir_id"] = jID(v, "rir")
		mapping["tenant_id"] = jID(v, "tenant")
		mapping["date_added"] = jStr(v, "date_added")
		mapping["description"] = jStr(v, "description")
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
	return d.Set("aggregates", s)
}
