package netbox

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxClusters() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxClustersRead,
		Description: `:meta:subcategory:Virtualization:`,
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
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"limit": {
				Type:             schema.TypeInt,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
			},
			"clusters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_type_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cluster_group_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"tenant_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"site_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"site_group_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"location_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"region_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"scope_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"scope_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"comments": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"custom_fields": {
							Type:     schema.TypeMap,
							Computed: true,
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
