package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxTenants() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxTenantsRead),
		Description: `:meta:subcategory:Tenancy:Retrieves all tenants from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
			"tenants": {
				Type:        schema.TypeList,
				Description: "The tenants of the tenants.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the tenants.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the tenants.",
							Computed:    true,
						},
						"slug": {
							Type:        schema.TypeString,
							Description: "The URL-friendly slug of the tenants.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the tenants.",
							Computed:    true,
						},
						"tags": tagsSchemaRead,
						"created": {
							Type:        schema.TypeString,
							Description: "The created of the tenants.",
							Computed:    true,
						},
						"last_updated": {
							Type:        schema.TypeString,
							Description: "The last updated of the tenants.",
							Computed:    true,
						},
						"comments": {
							Type:        schema.TypeString,
							Description: "Free-form comments about the tenants.",
							Computed:    true,
						},
						"custom_fields": {
							Type:        schema.TypeMap,
							Description: "Custom field values of the tenants, as a map of field name to value.",
							Computed:    true,
						},

						"site_count": {
							Type:        schema.TypeInt,
							Description: "The site count of the tenants.",
							Computed:    true,
						},
						"rack_count": {
							Type:        schema.TypeInt,
							Description: "The rack count of the tenants.",
							Computed:    true,
						},
						"device_count": {
							Type:        schema.TypeInt,
							Description: "The device count of the tenants.",
							Computed:    true,
						},
						"vrf_count": {
							Type:        schema.TypeInt,
							Description: "The vrf count of the tenants.",
							Computed:    true,
						},
						"prefix_count": {
							Type:        schema.TypeInt,
							Description: "The prefix count of the tenants.",
							Computed:    true,
						},
						"ip_address_count": {
							Type:        schema.TypeInt,
							Description: "The ip address count of the tenants.",
							Computed:    true,
						},
						"vlan_count": {
							Type:        schema.TypeInt,
							Description: "The vlan count of the tenants.",
							Computed:    true,
						},
						"vm_count": {
							Type:        schema.TypeInt,
							Description: "The vm count of the tenants.",
							Computed:    true,
						},
						"circuit_count": {
							Type:        schema.TypeInt,
							Description: "The circuit count of the tenants.",
							Computed:    true,
						},
						"cluster_count": {
							Type:        schema.TypeInt,
							Description: "The cluster count of the tenants.",
							Computed:    true,
						},
						"tenant_group": {
							Type:        schema.TypeList,
							Description: "The tenant group of the tenants.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Description: "The numeric ID of the tenants.",
										Computed:    true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "The name of the tenants.",
										Computed:    true,
									},
									"slug": {
										Type:        schema.TypeString,
										Description: "The URL-friendly slug of the tenants.",
										Computed:    true,
									},
									"tenant_count": {
										Type:        schema.TypeInt,
										Description: "The tenant count of the tenants.",
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

func dataSourceNetboxTenantsRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"name": "name",
		"slug": "slug",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "tenancy/tenants/", "tenants", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["name"] = jStr(v, "name")
		mapping["slug"] = jStr(v, "slug")
		mapping["description"] = jStr(v, "description")
		mapping["tags"] = jTags(v)
		mapping["created"] = jTime(v, "created")
		mapping["last_updated"] = jTime(v, "last_updated")
		mapping["comments"] = jStr(v, "comments")
		mapping["custom_fields"] = flattenCustomFields(v["custom_fields"])
		mapping["site_count"] = jInt(v, "site_count")
		mapping["rack_count"] = jInt(v, "rack_count")
		mapping["device_count"] = jInt(v, "device_count")
		mapping["vrf_count"] = jInt(v, "vrf_count")
		mapping["prefix_count"] = jInt(v, "prefix_count")
		mapping["ip_address_count"] = jInt(v, "ipaddress_count")
		mapping["vlan_count"] = jInt(v, "vlan_count")
		mapping["vm_count"] = jInt(v, "virtualmachine_count")
		mapping["circuit_count"] = jInt(v, "circuit_count")
		mapping["cluster_count"] = jInt(v, "cluster_count")
		mapping["tenant_group"] = jObjectList(v, "group", "id", "name", "slug", "tenant_count")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("tenants", s)
}
