package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxPrefix() *schema.Resource {
	filterAtLeastOneOf := []string{
		"description",
		"family",
		"prefix",
		"vlan_vid",
		"vrf_id",
		"vlan_id",
		"tenant_id",
		"site_id",
		"role_id",
		"cidr",
		"custom_fields",
		"tag",
		"status",
	}
	customFieldsFilterSchema := *customFieldsSchema
	customFieldsFilterSchema.AtLeastOneOf = filterAtLeastOneOf
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxPrefixRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves a single prefix from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Description: "The numeric ID of the prefix.",
				Computed:    true,
			},
			"cidr": {
				Type:          schema.TypeString,
				Description:   "The cidr of the prefix.",
				Optional:      true,
				Deprecated:    "The `cidr` parameter is deprecated in favor of the canonical `prefix` attribute.",
				ConflictsWith: []string{"prefix"},
				ValidateFunc:  validation.IsCIDR,
				AtLeastOneOf:  filterAtLeastOneOf,
			},
			customFieldsKey: &customFieldsFilterSchema,
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: filterAtLeastOneOf,
				Description:  "Description to include in the data source filter.",
			},
			"family": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: filterAtLeastOneOf,
				ValidateFunc: validation.IntInSlice([]int{4, 6}),
				Description:  "The IP family of the prefix. One of 4 or 6",
			},
			"role_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the role associated with the prefix.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: filterAtLeastOneOf,
			},
			"prefix": {
				Type:          schema.TypeString,
				Description:   "The prefix in CIDR notation.",
				Optional:      true,
				ValidateFunc:  validation.IsCIDR,
				ConflictsWith: []string{"cidr"},
				AtLeastOneOf:  filterAtLeastOneOf,
			},
			"vlan_vid": {
				Type:         schema.TypeFloat,
				Description:  "The vlan vid of the prefix.",
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
				ValidateFunc: validation.FloatBetween(1, 4094),
			},
			"vrf_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the vrf associated with the prefix.",
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
			},
			"vlan_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the vlan associated with the prefix.",
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
			},
			"tenant_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the tenant associated with the prefix.",
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
			},
			"site_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the site associated with the prefix.",
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
			},
			"site_group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the site group associated with the prefix.",
				Computed:    true,
			},
			"location_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the location associated with the prefix.",
				Computed:    true,
			},
			"region_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the region associated with the prefix.",
				Computed:    true,
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
Refer to [Netbox's documentation](https://demo.netbox.dev/static/docs/rest-api/filtering/#lookup-expressions)
for more information on available lookup expressions.`,
			},
			"status": {
				Type:         schema.TypeString,
				Description:  "The status of the prefix.",
				Optional:     true,
				AtLeastOneOf: filterAtLeastOneOf,
			},
			"tags": tagsSchemaRead,
		},
	}
}

type prefixDataSourceResponse struct {
	ID           int64                  `json:"id"`
	Prefix       string                 `json:"prefix"`
	Status       *netboxapi.Choice      `json:"status"`
	Description  string                 `json:"description"`
	Family       *netboxapi.IntChoice   `json:"family"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Role         *netboxapi.Ref         `json:"role"`
	Vrf          *netboxapi.Ref         `json:"vrf"`
	Vlan         *struct {
		ID  int64   `json:"id"`
		Vid float64 `json:"vid"`
	} `json:"vlan"`
	Tenant    *netboxapi.Ref `json:"tenant"`
	ScopeType *string        `json:"scope_type"`
	ScopeID   *int64         `json:"scope_id"`
}

func dataSourceNetboxPrefixRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	// note: cidr is deprecated in favor of prefix
	addStrQuery(query, d, "cidr", "prefix")
	addStrQuery(query, d, "description", "description")
	addIntQuery(query, d, "family", "family")
	addIntQuery(query, d, "role_id", "role_id")
	addStrQuery(query, d, "prefix", "prefix")
	addIntQuery(query, d, "vrf_id", "vrf_id")
	addIntQuery(query, d, "vlan_id", "vlan_id")
	if vlanVid, ok := d.Get("vlan_vid").(float64); ok && vlanVid != 0 {
		query.Set("vlan_vid", strconv.FormatFloat(vlanVid, 'f', -1, 64))
	}
	addIntQuery(query, d, "tenant_id", "tenant_id")
	addIntQuery(query, d, "site_id", "site_id")
	addStrQuery(query, d, "tag", "tag")
	addStrQuery(query, d, "tag__n", "tag__n")
	addStrQuery(query, d, "status", "status")
	if cfm, ok := d.Get(customFieldsKey).(map[string]interface{}); ok {
		addCustomFieldQuery(query, cfm)
	}

	result, err := findOne[prefixDataSourceResponse](m, "ipam/prefixes/", query, "prefix")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(result.ID, 10))
	d.Set("id", result.ID)
	d.Set("cidr", result.Prefix)
	d.Set("prefix", result.Prefix)
	d.Set("status", netboxapi.ChoiceValue(result.Status))
	d.Set("description", result.Description)
	if result.Family != nil {
		d.Set("family", int(result.Family.Value))
	}
	d.Set("tags", tagNames(result.Tags))
	d.Set(customFieldsKey, getCustomFields(result.CustomFields))
	if result.Role != nil {
		d.Set("role_id", result.Role.ID)
	}
	if result.Vrf != nil {
		d.Set("vrf_id", result.Vrf.ID)
	}
	if result.Vlan != nil {
		d.Set("vlan_vid", result.Vlan.Vid)
		d.Set("vlan_id", result.Vlan.ID)
	}
	if result.Tenant != nil {
		d.Set("tenant_id", result.Tenant.ID)
	}
	d.Set("site_id", nil)
	d.Set("site_group_id", nil)
	d.Set("location_id", nil)
	d.Set("region_id", nil)
	if result.ScopeType != nil && result.ScopeID != nil {
		switch *result.ScopeType {
		case "dcim.site":
			d.Set("site_id", *result.ScopeID)
		case "dcim.sitegroup":
			d.Set("site_group_id", *result.ScopeID)
		case "dcim.location":
			d.Set("location_id", *result.ScopeID)
		case "dcim.region":
			d.Set("region_id", *result.ScopeID)
		}
	}
	return nil
}
