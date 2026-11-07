package netbox

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxInterfaces() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxInterfaceRead,
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
			"limit": {
				Type:             schema.TypeInt,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
				Description:      "The limit of objects to return from the API lookup.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"interfaces": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mode": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type:     schema.TypeString,
								Computed: true,
							},
						},
						"mtu": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
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
						"tagged_vlans": {
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
								},
							},
						},
						// Do as a TypeList due to limitation of TypeMap
						"untagged_vlan": {
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
								},
							},
						},
						"vm_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxInterfaceRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"cluster_id":  "cluster_id",
		"mac_address": "mac_address",
		"name":        "name",
		"tag":         "tag",
		"vm_id":       "virtual_machine_id",
	}

	// Fetch everything and apply the regular expression before the limit
	var keep func(map[string]interface{}) bool
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		keep = func(o map[string]interface{}) bool { return r.MatchString(jStr(o, "name")) }
	}

	results, err := listFilteredWith[map[string]interface{}](d, m, "virtualization/interfaces/", "interfaces", allowed, nil, keep)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["description"] = jStr(v, "description")
		mapping["enabled"] = jBool(v, "enabled")
		mapping["mac_address"] = jStr(v, "mac_address")
		mapping["mode"] = jChoiceMap(v, "mode")
		mapping["mtu"] = jIntP(v, "mtu")
		mapping["name"] = jStr(v, "name")
		mapping["tagged_vlans"] = jObjectsList(v, "tagged_vlans", "id", "vid", "name")
		mapping["tag_ids"] = jTagIDs(v)
		mapping["untagged_vlan"] = jObjectList(v, "untagged_vlan", "id", "vid", "name")
		mapping["vm_id"] = jID(v, "virtual_machine")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("interfaces", s)
}
