package netbox

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxClusters() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxClustersRead),
		Description: `:meta:subcategory:Virtualization:Retrieves all clusters from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
				Description:  "The name regex of the clusters.",
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
			"clusters": {
				Type:        schema.TypeList,
				Description: "The clusters of the clusters.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the cluster associated with the clusters.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the clusters.",
							Computed:    true,
						},
						"cluster_type_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the cluster type associated with the clusters.",
							Computed:    true,
						},
						"cluster_group_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the cluster group associated with the clusters.",
							Computed:    true,
						},
						"tenant_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the tenant associated with the clusters.",
							Computed:    true,
						},
						"site_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the site associated with the clusters.",
							Computed:    true,
						},
						"site_group_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the site group associated with the clusters.",
							Computed:    true,
						},
						"location_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the location associated with the clusters.",
							Computed:    true,
						},
						"region_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the region associated with the clusters.",
							Computed:    true,
						},
						"scope_type": {
							Type:        schema.TypeString,
							Description: "The scope type of the clusters.",
							Computed:    true,
						},
						"scope_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the scope associated with the clusters.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the clusters.",
							Computed:    true,
						},
						"comments": {
							Type:        schema.TypeString,
							Description: "Free-form comments about the clusters.",
							Computed:    true,
						},
						"custom_fields": {
							Type:        schema.TypeMap,
							Description: "Custom field values of the clusters, as a map of field name to value.",
							Computed:    true,
						},
						tagsKey: tagsSchemaRead,
					},
				},
			},
		},
	}
}

func dataSourceNetboxClustersRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"name":             "name",
		"cluster_type_id":  "type_id",
		"cluster_group_id": "group_id",
		"site_id":          "site_id",
		"tag":              "tag",
	}

	// Fetch everything and apply the regular expression before the limit
	var keep func(map[string]interface{}) bool
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		keep = func(o map[string]interface{}) bool { return r.MatchString(jStr(o, "name")) }
	}

	results, err := listFilteredWith[map[string]interface{}](d, m, "virtualization/clusters/", "clusters", allowed, nil, keep)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["cluster_id"] = jInt(v, "id")
		mapping["name"] = jStr(v, "name")
		mapping["cluster_type_id"] = jID(v, "type")
		mapping["cluster_group_id"] = jID(v, "group")
		mapping["tenant_id"] = jID(v, "tenant")
		mapping["description"] = jStr(v, "description")
		mapping["comments"] = jStr(v, "comments")
		if scopeType := jStr(v, "scope_type"); scopeType != "" && v["scope_id"] != nil {
			scopeID := jIntP(v, "scope_id")
			mapping["scope_type"] = scopeType
			mapping["scope_id"] = scopeID
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
		mapping["custom_fields"] = v["custom_fields"]
		mapping[tagsKey] = jTags(v)

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("clusters", s)
}
