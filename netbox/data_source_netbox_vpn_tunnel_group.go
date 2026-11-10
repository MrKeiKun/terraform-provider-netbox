package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxVpnTunnelGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxVpnTunnelGroupRead),
		Description: `:meta:subcategory:VPN Tunnels:Retrieves a single vpn tunnel group from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Description: "The numeric ID of the vpn tunnel group.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the vpn tunnel group.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the vpn tunnel group.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the vpn tunnel group.",
				Optional:    true,
				Computed:    true,
			},
			"tags": tagsSchemaRead,
		},
	}
}

type vpnTunnelGroupDataSourceResponse struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func dataSourceNetboxVpnTunnelGroupRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "slug", "slug")
	addStrQuery(query, d, "tag", "tag")
	r, err := findOne[vpnTunnelGroupDataSourceResponse](m, "vpn/tunnel-groups/", query, "tunnel group")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("id", r.ID)
	d.Set("name", r.Name)
	d.Set("slug", r.Slug)
	d.Set("description", r.Description)
	d.Set("tags", tagNames(r.Tags))
	return nil
}
