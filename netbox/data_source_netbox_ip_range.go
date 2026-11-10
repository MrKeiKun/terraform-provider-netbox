package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxIPRange() *schema.Resource {
	filterAtLeastOneOf := []string{
		"contains",
		"family",
		"vrf_id",
		"tenant_id",
		"status",
		"role_id",
		"description",
		"tag",
	}
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxIPRangeRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves a single ip range from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Description: "The numeric ID of the ip range.",
				Computed:    true,
			},
			"start_address": {
				Type:        schema.TypeString,
				Description: "The start address of the ip range.",
				Computed:    true,
			},
			"end_address": {
				Type:        schema.TypeString,
				Description: "The end address of the ip range.",
				Computed:    true,
			},
			"contains": {
				Type:         schema.TypeString,
				Description:  "The contains of the ip range.",
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
				ValidateFunc: validation.IsCIDR,
			},
			"family": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: filterAtLeastOneOf,
				ValidateFunc: validation.IntInSlice([]int{4, 6}),
				Description:  "The IP family of the IP range. One of 4 or 6",
			},
			"vrf_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the vrf associated with the ip range.",
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
			},
			"tenant_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the tenant associated with the ip range.",
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
			},
			"status": {
				Type:         schema.TypeString,
				Description:  "The status of the ip range.",
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
			},
			"role_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the role associated with the ip range.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: filterAtLeastOneOf,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: filterAtLeastOneOf,
				Description:  "Description to include in the data source filter.",
			},
			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
				Description:  "Tag to include in the data source filter (must match the tag's slug).",
			},
			"tag__n": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `Tag to exclude from the data source filter (must match the tag's slug).
Refer to [Netbox's documentation](https://netboxlabs.com/docs/netbox/reference/filtering/#lookup-expressions)
for more information on available lookup expressions.`,
			},
			"tags": tagsSchemaRead,
		},
	}
}

type ipRangeDataSourceResponse struct {
	ID           int64                `json:"id"`
	StartAddress string               `json:"start_address"`
	EndAddress   string               `json:"end_address"`
	Status       *netboxapi.Choice    `json:"status"`
	Description  string               `json:"description"`
	Family       *netboxapi.IntChoice `json:"family"`
	Tags         []netboxapi.TagRef   `json:"tags"`
	Role         *netboxapi.Ref       `json:"role"`
	Vrf          *netboxapi.Ref       `json:"vrf"`
	Tenant       *netboxapi.Ref       `json:"tenant"`
}

func dataSourceNetboxIPRangeRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "contains", "contains")
	addIntQuery(query, d, "family", "family")
	addIntQuery(query, d, "vrf_id", "vrf_id")
	addIntQuery(query, d, "tenant_id", "tenant_id")
	addStrQuery(query, d, "status", "status")
	addIntQuery(query, d, "role_id", "role_id")
	addStrQuery(query, d, "description", "description")
	addStrQuery(query, d, "tag", "tag")
	addStrQuery(query, d, "tag__n", "tag__n")

	result, err := findOne[ipRangeDataSourceResponse](m, "ipam/ip-ranges/", query, "IP range")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(result.ID, 10))
	d.Set("id", result.ID)
	d.Set("start_address", result.StartAddress)
	d.Set("end_address", result.EndAddress)
	d.Set("status", netboxapi.ChoiceValue(result.Status))
	d.Set("description", result.Description)
	if result.Family != nil {
		d.Set("family", int(result.Family.Value))
	}
	d.Set("tags", tagNames(result.Tags))
	if result.Role != nil {
		d.Set("role_id", result.Role.ID)
	}
	if result.Vrf != nil {
		d.Set("vrf_id", result.Vrf.ID)
	}
	if result.Tenant != nil {
		d.Set("tenant_id", result.Tenant.ID)
	}
	return nil
}
