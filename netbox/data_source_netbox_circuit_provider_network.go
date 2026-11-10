package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxCircuitProviderNetwork() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxCircuitProviderNetworkRead),
		Description: `:meta:subcategory:Circuits:Retrieves a single circuit provider network from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"comments": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "Free-form comments about the circuit provider network.",
			},
			"custom_fields": {
				Computed:    true,
				Type:        schema.TypeMap,
				Description: "Custom field values of the circuit provider network, as a map of field name to value.",
			},
			"description": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "A description of the circuit provider network.",
			},
			"name": {
				AtLeastOneOf: []string{"id", "name"},
				Optional:     true,
				Type:         schema.TypeString,
				Description:  "The name of the circuit provider network.",
			},
			"id": {
				AtLeastOneOf: []string{"id", "name"},
				Computed:     true,
				Optional:     true,
				Type:         schema.TypeString,
				Description:  "The numeric ID of the circuit provider network.",
			},
			"provider_id": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "The ID of the provider associated with the circuit provider network.",
			},
			"service_id": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "The ID of the service associated with the circuit provider network.",
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

type circuitProviderNetworkDataSourceResponse struct {
	ID           int64                  `json:"id"`
	Comments     string                 `json:"comments"`
	Description  string                 `json:"description"`
	Name         string                 `json:"name"`
	Provider     *netboxapi.Ref         `json:"provider"`
	ServiceID    string                 `json:"service_id"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func dataSourceNetboxCircuitProviderNetworkRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "id", "id")

	result, err := findOneWith[circuitProviderNetworkDataSourceResponse](m, "circuits/provider-networks/", query,
		"more than one result, specify a more narrow filter", "no Provider Network found")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(result.ID, 10))
	d.Set("comments", result.Comments)
	d.Set("description", result.Description)
	d.Set("name", result.Name)
	if result.Provider != nil {
		d.Set("provider_id", result.Provider.ID)
	}
	d.Set("service_id", result.ServiceID)
	if result.CustomFields != nil {
		d.Set("custom_fields", flattenCustomFields(result.CustomFields))
	}
	d.Set(tagsKey, tagNames(result.Tags))
	return nil
}
