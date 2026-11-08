package netbox

import (
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxVlans() *schema.Resource {
	customFieldsFilterSchema := *customFieldsSchema
	return &schema.Resource{
		Read:        dataSourceNetboxVlansRead,
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
			customFieldsKey: &customFieldsFilterSchema,
			"vlans": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"site": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tag_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"tenant": {
							Type:     schema.TypeInt,
							Computed: true,
						},
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

func dataSourceNetboxVlansRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"vid":                "vid",
		"vid__gt":            "vid__gt",
		"vid__gte":           "vid__gte",
		"vid__lt":            "vid__lt",
		"vid__lte":           "vid__lte",
		"vid__n":             "vid__n",
		"group":              "group",
		"group__n":           "group__n",
		"group_id":           "group_id",
		"group_id__n":        "group_id__n",
		"tag":                "tag",
		"tenant":             "tenant",
		"tenant__n":          "tenant__n",
		"tenant_group":       "tenant_group",
		"tenant_group__n":    "tenant_group__n",
		"tenant_group_id":    "tenant_group_id",
		"tenant_group_id__n": "tenant_group_id__n",
		"tenant_id":          "tenant_id",
		"tenant_id__n":       "tenant_id__n",
		"site":               "site",
		"site__n":            "site__n",
		"site_id":            "site_id",
		"site_id__n":         "site_id__n",
		"status":             "status",
	}

	extra := url.Values{}
	if cfm, ok := d.Get(customFieldsKey).(map[string]interface{}); ok {
		addCustomFieldQuery(extra, cfm)
	}

	results, err := listFiltered[map[string]interface{}](d, m, "ipam/vlans/", "VLANs", allowed, extra)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["vid"] = jIntP(v, "vid")
		mapping["name"] = jStr(v, "name")
		mapping["description"] = jStr(v, "description")
		mapping["group_id"] = jID(v, "group")
		mapping["vid"] = jIntP(v, "vid")
		mapping["role"] = jID(v, "role")
		mapping["site"] = jID(v, "site")
		mapping["status"] = jChoice(v, "status")
		mapping["tenant"] = jID(v, "tenant")
		if cf := flattenCustomFields(v["custom_fields"]); cf != nil {
			mapping[customFieldsKey] = cf
		}
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
	return d.Set("vlans", s)
}
