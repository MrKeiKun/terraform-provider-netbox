package netbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxAvailablePrefix() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxAvailablePrefixRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves a single available prefix from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"prefix_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the prefix associated with the available prefix.",
				Required:    true,
			},
			"prefixes_available": {
				Type:        schema.TypeList,
				Description: "The prefixes available of the available prefix.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"family": {
							Type:        schema.TypeInt,
							Description: "The IP family, 4 or 6.",
							Computed:    true,
						},
						"prefix": {
							Type:        schema.TypeString,
							Description: "The prefix in CIDR notation.",
							Computed:    true,
						},
						"vrf_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the vrf associated with the available prefix.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

type availablePrefixDataSourceResponse struct {
	Prefix string         `json:"prefix"`
	Family float64        `json:"family"`
	Vrf    *netboxapi.Ref `json:"vrf"`
}

func dataSourceNetboxAvailablePrefixRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	var prefixID int64
	if v, ok := d.Get("prefix_id").(int); ok && v != 0 {
		prefixID = int64(v)
	}

	path := fmt.Sprintf("ipam/prefixes/%d/available-prefixes/", prefixID)
	result, err := netboxapi.Request[[]availablePrefixDataSourceResponse](context.Background(), api.rest, http.MethodGet, path, nil, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range *result {
		mapping := make(map[string]interface{})
		mapping["prefix"] = v.Prefix
		mapping["family"] = v.Family
		if v.Vrf != nil {
			mapping["vrf_id"] = v.Vrf.ID
		}
		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("prefixes_available", s)
}
