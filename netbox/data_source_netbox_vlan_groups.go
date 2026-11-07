package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxVlanGroups() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxVlanGroupsRead,
		Description: `:meta:subcategory:IP Address Management (IPAM):`,
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
			"vlan_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slug": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ranges": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"end": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"used": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"tag_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxVlanGroupsRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"name":               "name",
		"name__empty":        "name__empty",
		"name__ic":           "name__ic",
		"name__ie":           "name__ie",
		"name__iew":          "name__iew",
		"name__isw":          "name__isw",
		"name__n":            "name__n",
		"name__nic":          "name__nic",
		"name__nie":          "name__nie",
		"name__niew":         "name__niew",
		"name__nisw":         "name__nisw",
		"slug":               "slug",
		"slug__empty":        "slug__empty",
		"slug__ic":           "slug__ic",
		"slug__ie":           "slug__ie",
		"slug__iew":          "slug__iew",
		"slug__isw":          "slug__isw",
		"slug__n":            "slug__n",
		"slug__nic":          "slug__nic",
		"slug__nie":          "slug__nie",
		"slug__niew":         "slug__niew",
		"slug__nisw":         "slug__nisw",
		"description":        "description",
		"description__empty": "description__empty",
		"description__ic":    "description__ic",
		"description__ie":    "description__ie",
		"description__iew":   "description__iew",
		"description__isw":   "description__isw",
		"description__n":     "description__n",
		"description__nic":   "description__nic",
		"description__nie":   "description__nie",
		"description__niew":  "description__niew",
		"description__nisw":  "description__nisw",
		"id":                 "id",
		"id__gt":             "id__gt",
		"id__gte":            "id__gte",
		"id__lt":             "id__lt",
		"id__lte":            "id__lte",
		"id__n":              "id__n",
		"scope_type":         "scope_type",
		"scope_type__n":      "scope_type__n",
		"scope_id":           "scope_id",
		"scope_id__gt":       "scope_id__gt",
		"scope_id__gte":      "scope_id__gte",
		"scope_id__lt":       "scope_id__lt",
		"scope_id__lte":      "scope_id__lte",
		"scope_id__n":        "scope_id__n",
		"site_id":            "site",
		"site":               "site",
		"location_id":        "location",
		"location":           "location",
		"rack_id":            "rack",
		"rack":               "rack",
		"region_id":          "region",
		"region":             "region",
		"sitegroup_id":       "sitegroup",
		"sitegroup":          "sitegroup",
		"cluster_id":         "cluster",
		"cluster":            "cluster",
		"clustergroup_id":    "clustergroup",
		"clustergroup":       "clustergroup",
		"minvid":             "min_vid",
		"minvid__gt":         "min_vid__gt",
		"minvid__gte":        "min_vid__gte",
		"minvid__lt":         "min_vid__lt",
		"minvid__lte":        "min_vid__lte",
		"minvid__n":          "min_vid__n",
		"maxvid":             "max_vid",
		"maxvid__gt":         "max_vid__gt",
		"maxvid__gte":        "max_vid__gte",
		"maxvid__lt":         "max_vid__lt",
		"maxvid__lte":        "max_vid__lte",
		"maxvid__n":          "max_vid__n",
		"tag":                "tag",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "ipam/vlan-groups/", "VLAN groups", allowed, nil)
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
		mapping["ranges"] = jRanges(v, "vid_ranges")
		mapping["used"] = jInt(v, "vlan_count")
		mapping["tag_ids"] = jTagIDs(v)

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("vlan_groups", s)
}
