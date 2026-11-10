package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxIPRanges() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxIPRangesRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves all ip ranges from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
			"ip_ranges": {
				Type:        schema.TypeList,
				Description: "The ip ranges of the ip ranges.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the ip ranges.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the ip ranges.",
							Computed:    true,
						},
						"created": {
							Type:        schema.TypeString,
							Description: "The created of the ip ranges.",
							Computed:    true,
						},
						"last_updated": {
							Type:        schema.TypeString,
							Description: "The last updated of the ip ranges.",
							Computed:    true,
						},
						"custom_fields": {
							Type:        schema.TypeMap,
							Description: "Custom field values of the ip ranges, as a map of field name to value.",
							Computed:    true,
						},
						"start_address": {
							Type:        schema.TypeString,
							Description: "The start address of the ip ranges.",
							Computed:    true,
						},
						"end_address": {
							Type:        schema.TypeString,
							Description: "The end address of the ip ranges.",
							Computed:    true,
						},
						"address_family": {
							Type:        schema.TypeString,
							Description: "The address family of the ip ranges.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "The status of the ip ranges.",
							Computed:    true,
						},
						"role": {
							Type:        schema.TypeString,
							Description: "The role of the ip ranges.",
							Computed:    true,
						},

						"tenant": {
							Type:        schema.TypeList,
							Description: "The tenant of the ip ranges.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Description: "The numeric ID of the ip ranges.",
										Computed:    true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "The name of the ip ranges.",
										Computed:    true,
									},
									"slug": {
										Type:        schema.TypeString,
										Description: "The URL-friendly slug of the ip ranges.",
										Computed:    true,
									},
								},
							},
						},
						"tags": {
							Type:        schema.TypeList,
							Description: "Tags to assign to the ip ranges.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Description: "The numeric ID of the ip ranges.",
										Computed:    true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "The name of the ip ranges.",
										Computed:    true,
									},
									"display": {
										Type:        schema.TypeString,
										Description: "The display of the ip ranges.",
										Computed:    true,
									},
									"slug": {
										Type:        schema.TypeString,
										Description: "The URL-friendly slug of the ip ranges.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxIPRangesRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"contains":      "contains",
		"start_address": "start_address",
		"end_address":   "end_address",
		"role":          "role",
		"status":        "status",
		"vrf":           "vrf",
		"tenant":        "tenant",
		"tag":           "tag",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "ipam/ip-ranges/", "IP ranges", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["description"] = jStr(v, "description")
		mapping["created"] = jTime(v, "created")
		mapping["last_updated"] = jTime(v, "last_updated")
		mapping["custom_fields"] = flattenCustomFields(v["custom_fields"])
		mapping["start_address"] = jStr(v, "start_address")
		mapping["end_address"] = jStr(v, "end_address")
		mapping["address_family"] = jNested(v, "family", "label")
		mapping["status"] = jChoice(v, "status")
		mapping["tenant"] = jObjectList(v, "tenant", "id", "name", "slug")
		mapping["vrf_id"] = jID(v, "vrf")
		mapping["role_id"] = jID(v, "role")
		mapping["description"] = jStr(v, "description")
		mapping["tags"] = jObjectsList(v, "tags", "name", "display", "slug", "id")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("ip_ranges", s)
}
