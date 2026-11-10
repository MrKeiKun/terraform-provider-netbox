package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxVrfs() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxVrfsRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves all vrfs from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
			"vrfs": {
				Type:        schema.TypeList,
				Description: "The vrfs of the vrfs.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the vrfs.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the vrfs.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the vrfs.",
							Computed:    true,
						},
						"rd": {
							Type:        schema.TypeString,
							Description: "The route distinguisher of the vrfs.",
							Computed:    true,
						},
						"tenant": {
							Type:        schema.TypeInt,
							Description: "The tenant of the vrfs.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxVrfsRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"id":                 "id",
		"name":               "name",
		"description":        "description",
		"rd":                 "rd",
		"tenant":             "tenant",
		"tenant__n":          "tenant__n",
		"tenant_group":       "tenant_group",
		"tenant_group__n":    "tenant_group__n",
		"tenant_group_id":    "tenant_group_id",
		"tenant_group_id__n": "tenant_group_id__n",
		"tenant_id":          "tenant_id",
		"tenant_id__n":       "tenant_id__n",
		"tag":                "tag",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "ipam/vrfs/", "VRFs", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["name"] = jStr(v, "name")
		mapping["description"] = jStr(v, "description")
		mapping["rd"] = jStr(v, "rd")
		mapping["tenant"] = jID(v, "tenant")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("vrfs", s)
}
