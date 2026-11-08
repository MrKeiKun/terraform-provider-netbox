package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxPrefixes() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxPrefixesRead,
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves all prefixes from NetBox. Narrow the result with filter blocks and cap it with limit.`,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A list of filters to apply to the API query when requesting prefixes.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the field to filter on. Supported fields are: `prefix`, `contains`, `vlan_vid`, `vrf_id`, `vlan_id`, `status`, `tenant_id`, `site_id`, `role_id`, `region_id`, `description` & `tag`.",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The value to pass to the specified filter.",
						},
					},
				},
			},
			"limit": {
				Type:             schema.TypeInt,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
				Description:      "The limit of objects to return from the API lookup.",
			},
			"prefixes": {
				Type:        schema.TypeList,
				Description: "The prefixes of the prefixes.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the prefixes.",
							Computed:    true,
						},
						"prefix": {
							Type:        schema.TypeString,
							Description: "The prefix in CIDR notation.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the prefixes.",
							Computed:    true,
						},
						"tenant_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the tenant associated with the prefixes.",
							Computed:    true,
						},
						"site_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the site associated with the prefixes.",
							Computed:    true,
						},
						"site_group_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the site group associated with the prefixes.",
							Computed:    true,
						},
						"location_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the location associated with the prefixes.",
							Computed:    true,
						},
						"region_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the region associated with the prefixes.",
							Computed:    true,
						},
						"role_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the role associated with the prefixes.",
							Computed:    true,
						},
						"vlan_vid": {
							Type:        schema.TypeFloat,
							Description: "The vlan vid of the prefixes.",
							Computed:    true,
						},
						"vrf_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the vrf associated with the prefixes.",
							Computed:    true,
						},
						"vlan_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the vlan associated with the prefixes.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "The status of the prefixes.",
							Computed:    true,
						},
						"tags": tagsSchemaRead,
						customFieldsKey: {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxPrefixesRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"prefix":      "prefix",
		"vlan_vid":    "vlan_vid",
		"contains":    "contains",
		"vrf_id":      "vrf_id",
		"vlan_id":     "vlan_id",
		"status":      "status",
		"tenant_id":   "tenant_id",
		"site_id":     "site_id",
		"role_id":     "role_id",
		"region_id":   "region_id",
		"description": "description",
		"tag":         "tag",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "ipam/prefixes/", "prefixes", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["prefix"] = jStr(v, "prefix")
		mapping["description"] = jStr(v, "description")
		mapping["vlan_vid"] = jNested(v, "vlan", "vid")
		mapping["vlan_id"] = jID(v, "vlan")
		mapping["vrf_id"] = jID(v, "vrf")
		mapping["tenant_id"] = jID(v, "tenant")
		mapping["role_id"] = jID(v, "role")
		if scopeType := jStr(v, "scope_type"); scopeType != "" && v["scope_id"] != nil {
			scopeID := jInt(v, "scope_id")
			switch scopeType {
			case "dcim.site":
				mapping["site_id"] = scopeID
			case "dcim.sitegroup":
				mapping["site_group_id"] = scopeID
			case "dcim.location":
				mapping["location_id"] = scopeID
			case "dcim.region":
				mapping["region_id"] = scopeID
			}
		}
		mapping["status"] = jChoice(v, "status")
		mapping["tags"] = jTags(v)
		if cf := flattenCustomFields(v["custom_fields"]); cf != nil {
			mapping[customFieldsKey] = cf
		}

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("prefixes", s)
}
