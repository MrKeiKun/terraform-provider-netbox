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
		ReadContext: withContext(dataSourceNetboxIPAddressRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves a single ip address from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeInt,
				Description:  "The numeric ID of the ip address.",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"ip_address": {
				Type:        schema.TypeString,
				Description: "The IP address of the ip address, in CIDR notation.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ip address.",
				Computed:    true,
			},
			"created": {
				Type:        schema.TypeString,
				Description: "The created of the ip address.",
				Computed:    true,
			},
			"last_updated": {
				Type:        schema.TypeString,
				Description: "The last updated of the ip address.",
				Computed:    true,
			},
			"custom_fields": {
				Type:        schema.TypeMap,
				Description: "Custom field values of the ip address, as a map of field name to value.",
				Computed:    true,
			},
			"address_family": {
				Type:        schema.TypeString,
				Description: "The address family of the ip address.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the ip address.",
				Computed:    true,
			},
			"dns_name": {
				Type:        schema.TypeString,
				Description: "The DNS name of the ip address.",
				Computed:    true,
			},
			"role": {
				Type:        schema.TypeString,
				Description: "The role of the ip address.",
				Computed:    true,
			},

			"tenant": {
				Type:        schema.TypeList,
				Description: "The tenant of the ip address.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the ip address.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the ip address.",
							Computed:    true,
						},
						"slug": {
							Type:        schema.TypeString,
							Description: "The URL-friendly slug of the ip address.",
							Computed:    true,
						},
					},
				},
			},
			"tags": {
				Type:        schema.TypeList,
				Description: "Tags to assign to the ip address.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the ip address.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the ip address.",
							Computed:    true,
						},
						"display": {
							Type:        schema.TypeString,
							Description: "The display of the ip address.",
							Computed:    true,
						},
						"slug": {
							Type:        schema.TypeString,
							Description: "The URL-friendly slug of the ip address.",
							Computed:    true,
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
