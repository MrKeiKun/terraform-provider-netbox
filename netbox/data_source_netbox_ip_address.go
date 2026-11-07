package netbox

import (
	"context"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxIPAddress() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxIPAddressRead,
		Description: `:meta:subcategory:IP Address Management (IPAM):`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"ip_address": {
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
			"custom_fields": {
				Type:     schema.TypeMap,
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
	}
}

type ipAddressDataSourceResponse struct {
	ID           int64                  `json:"id"`
	Address      string                 `json:"address"`
	Description  string                 `json:"description"`
	Created      string                 `json:"created"`
	LastUpdated  string                 `json:"last_updated"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Family       *struct {
		Label string `json:"label"`
	} `json:"family"`
	Status  *netboxapi.Choice `json:"status"`
	DNSName string            `json:"dns_name"`
	Role    *netboxapi.Choice `json:"role"`
	Tenant  *struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"tenant"`
	Tags []struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		Display string `json:"display"`
		Slug    string `json:"slug"`
	} `json:"tags"`
}

func dataSourceNetboxIPAddressRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	id := d.Get("id").(int)
	result, err := netboxapi.Get[ipAddressDataSourceResponse](context.Background(), api.rest, "ipam/ip-addresses/", int64(id))
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(result.ID, 10))
	d.Set("id", result.ID)
	d.Set("ip_address", result.Address)
	d.Set("description", result.Description)
	d.Set("created", jTime(map[string]interface{}{"created": result.Created}, "created"))
	d.Set("last_updated", jTime(map[string]interface{}{"last_updated": result.LastUpdated}, "last_updated"))
	d.Set("custom_fields", flattenCustomFields(result.CustomFields))
	if result.Family != nil {
		d.Set("address_family", result.Family.Label)
	}
	d.Set("status", netboxapi.ChoiceValue(result.Status))
	d.Set("dns_name", result.DNSName)
	if result.Role != nil {
		d.Set("role", result.Role.Value)
	}

	var tenant []map[string]interface{}
	if result.Tenant != nil {
		tenant = append(tenant, map[string]interface{}{
			"id":   result.Tenant.ID,
			"name": result.Tenant.Name,
			"slug": result.Tenant.Slug,
		})
	}
	d.Set("tenant", tenant)

	var tags []map[string]interface{}
	for _, t := range result.Tags {
		tags = append(tags, map[string]interface{}{
			"name":    t.Name,
			"display": t.Display,
			"slug":    t.Slug,
			"id":      t.ID,
		})
	}
	d.Set("tags", tags)
	return nil
}
