package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxVpnTunnel() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxVpnTunnelRead),
		Description: `:meta:subcategory:VPN Tunnels:Retrieves a single vpn tunnel from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Description: "The numeric ID of the vpn tunnel.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the vpn tunnel.",
				Required:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the vpn tunnel.",
				Optional:    true,
				Computed:    true,
			},
			"encapsualation": {
				Type:        schema.TypeString,
				Description: "The encapsualation of the vpn tunnel.",
				Optional:    true,
				Computed:    true,
			},
			"tunnel_group_id": {
				Type:        schema.TypeString,
				Description: "The ID of the tunnel group associated with the vpn tunnel.",
				Optional:    true,
				Computed:    true,
			},
			"tunnel_id": {
				Type:        schema.TypeString,
				Description: "The ID of the tunnel associated with the vpn tunnel.",
				Optional:    true,
				Computed:    true,
			},
			"tenant_id": {
				Type:        schema.TypeString,
				Description: "The ID of the tenant associated with the vpn tunnel.",
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the vpn tunnel.",
				Computed:    true,
			},
			"tags": tagsSchemaRead,
		},
	}
}

type vpnTunnelDataSourceResponse struct {
	ID            int64              `json:"id"`
	Name          string             `json:"name"`
	Status        *netboxapi.Choice  `json:"status"`
	Encapsulation *netboxapi.Choice  `json:"encapsulation"`
	Group         *netboxapi.Ref     `json:"group"`
	TunnelID      *int64             `json:"tunnel_id"`
	Tenant        *netboxapi.Ref     `json:"tenant"`
	Description   string             `json:"description"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

func dataSourceNetboxVpnTunnelRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "status", "status")
	addStrQuery(query, d, "encapsulation", "encapsulation")
	addStrQuery(query, d, "group_id", "group_id")
	addStrQuery(query, d, "tunnel_id", "tunnel_id")
	addStrQuery(query, d, "tenant_id", "tenant_id")
	addStrQuery(query, d, "tag", "tag")

	r, err := findOne[vpnTunnelDataSourceResponse](m, "vpn/tunnels/", query, "vpn tunnel")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("id", r.ID)
	d.Set("name", r.Name)
	d.Set("status", netboxapi.ChoiceValue(r.Status))
	d.Set("encapsualation", netboxapi.ChoiceValue(r.Encapsulation))
	if r.Group != nil {
		d.Set("tunnel_group_id", strconv.FormatInt(r.Group.ID, 10))
	}
	if r.TunnelID != nil {
		d.Set("tunnel_id", strconv.FormatInt(*r.TunnelID, 10))
	}
	if r.Tenant != nil {
		d.Set("tenant_id", strconv.FormatInt(r.Tenant.ID, 10))
	}
	d.Set("description", r.Description)
	d.Set("tags", tagNames(r.Tags))
	return nil
}
