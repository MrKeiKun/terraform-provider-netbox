package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxVpnTunnel() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxVpnTunnelRead,
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
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"encapsualation": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tunnel_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tunnel_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
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
