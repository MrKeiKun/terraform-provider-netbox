package netbox

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxCircuitProviderNetworks() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxCircuitProviderNetworksRead,
		Description: `:meta:subcategory:Circuits:`,
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
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
				Default:          0,
			},
			"provider_networks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comments": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"custom_fields": {
							Type:     schema.TypeMap,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"provider_network_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"provider_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"service_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						tagsKey: tagsSchemaRead,
					},
				},
			},
		},
	}
}

func dataSourceNetboxCircuitProviderNetworksRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"name":        "name",
		"provider_id": "provider_id",
		"tag":         "tag",
	}

	// Fetch everything and apply the regular expression before the limit
	var keep func(map[string]interface{}) bool
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		keep = func(o map[string]interface{}) bool { return r.MatchString(jStr(o, "name")) }
	}

	results, err := listFilteredWith[map[string]interface{}](d, m, "circuits/provider-networks/", "Provider Networks", allowed, nil, keep)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["provider_network_id"] = jInt(v, "id")
		mapping["comments"] = jStr(v, "comments")
		mapping["custom_fields"] = v["custom_fields"]
		mapping["description"] = jStr(v, "description")
		mapping["name"] = jStr(v, "name")
		mapping["provider_id"] = jID(v, "provider")
		mapping["service_id"] = jStr(v, "service_id")
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
	return d.Set("provider_networks", s)
}
