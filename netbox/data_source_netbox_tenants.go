package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxTenants() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxTenantsRead,
		Description: `:meta:subcategory:Tenancy:`,
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
			"tenants": {
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
						"created": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_updated": {
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

						"site_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"rack_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"device_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vrf_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"prefix_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ip_address_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vlan_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vm_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"circuit_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cluster_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"tenant_group": {
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
									"tenant_count": {
										Type:     schema.TypeInt,
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
