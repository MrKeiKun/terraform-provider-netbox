package netbox

import (
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxIPAddresses() *schema.Resource {
	customFieldsFilterSchema := *customFieldsSchema
	return &schema.Resource{
		Read:        dataSourceNetboxIPAddressesRead,
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
			"ip_addresses": {
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
						"created": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_updated": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"custom_fields": {
							Type:     schema.TypeMap,
							Computed: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dns_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"assigned_object_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Content type of the object the address is assigned to, for example dcim.interface.",
						},
						"assigned_object_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "ID of the object the address is assigned to.",
						},

						"tenant": {
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
								},
							},
						},
						"tags": {
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
									"display": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"slug": {
										Type:     schema.TypeString,
										Computed: true,
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

func dataSourceNetboxIPAddressesRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"dns_name":        "dns_name",
		"interface_id":    "interface_id",
		"device_id":       "device_id",
		"ip_address":      "address",
		"vm_interface_id": "vminterface_id",
		"role":            "role",
		"status":          "status",
		"vrf":             "vrf",
		"tenant":          "tenant",
		"parent_prefix":   "parent",
		"tag":             "tag",
		"description":     "description",
	}

	extra := url.Values{}
	if cfm, ok := d.Get(customFieldsKey).(map[string]interface{}); ok {
		addCustomFieldQuery(extra, cfm)
	}

	results, err := listFiltered[map[string]interface{}](d, m, "ipam/ip-addresses/", "IP addresses", allowed, extra)
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
		mapping["ip_address"] = jStr(v, "address")
		mapping["address_family"] = jNested(v, "family", "label")
		mapping["status"] = jChoice(v, "status")
		mapping["dns_name"] = jStr(v, "dns_name")
		mapping["tenant"] = jObjectList(v, "tenant", "id", "name", "slug")
		mapping["tags"] = jObjectsList(v, "tags", "name", "display", "slug", "id")
		mapping["role"] = jChoice(v, "role")
		mapping["assigned_object_type"] = jStr(v, "assigned_object_type")
		mapping["assigned_object_id"] = jInt(v, "assigned_object_id")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("ip_addresses", s)
}
