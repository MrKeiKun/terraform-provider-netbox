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
		Read:        dataSourceNetboxVpnTunnelGroupRead,
		Description: `:meta:subcategory:VPN Tunnels:`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
