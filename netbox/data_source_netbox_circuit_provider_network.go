package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxCircuitProviderNetwork() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxCircuitProviderNetworkRead,
		Description: `:meta:subcategory:Circuits:`,
		Schema: map[string]*schema.Schema{
			"comments": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"custom_fields": {
				Computed: true,
				Type:     schema.TypeMap,
			},
			"description": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"name": {
				AtLeastOneOf: []string{"id", "name"},
				Optional:     true,
				Type:         schema.TypeString,
			},
			"id": {
				AtLeastOneOf: []string{"id", "name"},
				Computed:     true,
				Optional:     true,
				Type:         schema.TypeString,
			},
			"provider_id": {
				Computed: true,
				Type:     schema.TypeInt,
			},
			"service_id": {
				Computed: true,
				Type:     schema.TypeString,
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
