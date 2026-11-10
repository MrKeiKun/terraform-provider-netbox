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
		ReadContext: withContext(dataSourceNetboxIPAddressesRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves all ip addresses from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
			customFieldsKey: &customFieldsFilterSchema,
			"ip_addresses": {
				Type:        schema.TypeList,
				Description: "The ip addresses of the ip addresses.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the ip addresses.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the ip addresses.",
							Computed:    true,
						},
						"created": {
							Type:        schema.TypeString,
							Description: "The created of the ip addresses.",
							Computed:    true,
						},
						"last_updated": {
							Type:        schema.TypeString,
							Description: "The last updated of the ip addresses.",
							Computed:    true,
						},
						"custom_fields": {
							Type:        schema.TypeMap,
							Description: "Custom field values of the ip addresses, as a map of field name to value.",
							Computed:    true,
						},
						"ip_address": {
							Type:        schema.TypeString,
							Description: "The IP address of the ip addresses, in CIDR notation.",
							Computed:    true,
						},
						"address_family": {
							Type:        schema.TypeString,
							Description: "The address family of the ip addresses.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "The status of the ip addresses.",
							Computed:    true,
						},
						"dns_name": {
							Type:        schema.TypeString,
							Description: "The DNS name of the ip addresses.",
							Computed:    true,
						},
						"role": {
							Type:        schema.TypeString,
							Description: "The role of the ip addresses.",
							Computed:    true,
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
							Type:        schema.TypeList,
							Description: "The tenant of the ip addresses.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Description: "The numeric ID of the ip addresses.",
										Computed:    true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "The name of the ip addresses.",
										Computed:    true,
									},
									"slug": {
										Type:        schema.TypeString,
										Description: "The URL-friendly slug of the ip addresses.",
										Computed:    true,
									},
								},
							},
						},
						"tags": {
							Type:        schema.TypeList,
							Description: "Tags to assign to the ip addresses.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Description: "The numeric ID of the ip addresses.",
										Computed:    true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "The name of the ip addresses.",
										Computed:    true,
									},
									"display": {
										Type:        schema.TypeString,
										Description: "The display of the ip addresses.",
										Computed:    true,
									},
									"slug": {
										Type:        schema.TypeString,
										Description: "The URL-friendly slug of the ip addresses.",
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
