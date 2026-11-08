package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxVlanGroup() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxVlanGroupRead,
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves a single vlan group from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Description:  "The numeric ID of the vlan group.",
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"id", "name", "slug", "scope_type"},
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the vlan group.",
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"id", "name", "slug", "scope_type"},
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the vlan group.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "name", "slug", "scope_type"},
			},
			"scope_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxVlanGroupScopeTypeOptions, false),
				Description:  buildValidValueDescription(resourceNetboxVlanGroupScopeTypeOptions),
				AtLeastOneOf: []string{"id", "name", "slug", "scope_type"},
			},
			"scope_id": {
				Type:         schema.TypeString,
				Description:  "The ID of the scope associated with the vlan group.",
				Optional:     true,
				RequiredWith: []string{"scope_type"},
			},
			"vlan_count": {
				Type:        schema.TypeInt,
				Description: "The vlan count of the vlan group.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the vlan group.",
				Computed:    true,
			},
		},
	}
}

type vlanGroupDataSourceResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	VlanCount   int64   `json:"vlan_count"`
	Description string  `json:"description"`
	ScopeID     *int64  `json:"scope_id"`
	ScopeType   *string `json:"scope_type"`
}

func dataSourceNetboxVlanGroupRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "id", "id")
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "slug", "slug")
	addStrQuery(query, d, "scope_type", "scope_type")
	addStrQuery(query, d, "scope_id", "scope_id")

	r, err := findOne[vlanGroupDataSourceResponse](m, "ipam/vlan-groups/", query, "vlan group")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("id", strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	d.Set("slug", r.Slug)
	d.Set("vlan_count", r.VlanCount)
	d.Set("description", r.Description)
	if r.ScopeID != nil {
		d.Set("scope_id", strconv.FormatInt(*r.ScopeID, 10))
	}
	if r.ScopeType != nil {
		d.Set("scope_type", *r.ScopeType)
	}
	return nil
}
