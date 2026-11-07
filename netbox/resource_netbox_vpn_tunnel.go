package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var resourceNetboxVpnTunnelEncapsulationOptions = []string{"ipsec-transport", "ipsec-tunnel", "ip-ip", "gre"}
var resourceNetboxVpnTunnelStatusOptions = []string{"planned", "active", "disabled"}

func resourceNetboxVpnTunnel() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVpnTunnelCreate,
		Read:   resourceNetboxVpnTunnelRead,
		Update: resourceNetboxVpnTunnelUpdate,
		Delete: resourceNetboxVpnTunnelDelete,

		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/features/vpn-tunnels/):

> NetBox can model private tunnels formed among virtual termination points across your network. Typical tunnel implementations include GRE, IP-in-IP, and IPSec. A tunnel may be terminated to two or more device or virtual machine interfaces. For convenient organization, tunnels may be assigned to user-defined groups.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"encapsulation": {
				Type:        schema.TypeString,
				Required:    true,
				Description: buildValidValueDescription(resourceNetboxVpnTunnelEncapsulationOptions),
			},
			"status": {
				Type:        schema.TypeString,
				Required:    true,
				Description: buildValidValueDescription(resourceNetboxVpnTunnelStatusOptions),
			},
			"tunnel_group_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tunnel_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			tagsKey: tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const vpnTunnelAPIPath = "vpn/tunnels/"

type vpnTunnelRequest struct {
	Name          string             `json:"name"`
	Encapsulation string             `json:"encapsulation"`
	Status        string             `json:"status"`
	Group         *int64             `json:"group"`
	Description   string             `json:"description,omitempty"`
	Tenant        *int64             `json:"tenant,omitempty"`
	TunnelId      *int64             `json:"tunnel_id,omitempty"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

type vpnTunnelResponse struct {
	Name          string             `json:"name"`
	Encapsulation *netboxapi.Choice  `json:"encapsulation"`
	Status        *netboxapi.Choice  `json:"status"`
	Group         *netboxapi.Ref     `json:"group"`
	Description   string             `json:"description"`
	Tenant        *netboxapi.Ref     `json:"tenant"`
	TunnelId      *int64             `json:"tunnel_id"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

func vpnTunnelRequestFromResourceData(api *providerState, d *schema.ResourceData) (*vpnTunnelRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &vpnTunnelRequest{
		Name:          name,
		Encapsulation: d.Get("encapsulation").(string),
		Status:        d.Get("status").(string),
		Group:         int64ToPtr(int64(d.Get("tunnel_group_id").(int))),
		Description:   d.Get("description").(string),
		Tenant:        getOptionalInt(d, "tenant_id"),
		TunnelId:      getOptionalInt(d, "tunnel_id"),
		Tags:          tags,
	}
	return data, nil
}

func resourceNetboxVpnTunnelCreate(d *schema.ResourceData, m interface{}) error {
	data, err := vpnTunnelRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, vpnTunnelAPIPath, data, resourceNetboxVpnTunnelRead)
}

func resourceNetboxVpnTunnelRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, vpnTunnelAPIPath, func(r *vpnTunnelResponse) {
		d.Set("name", r.Name)
		d.Set("encapsulation", netboxapi.ChoiceValue(r.Encapsulation))
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("tunnel_group_id", netboxapi.RefID(r.Group))
		d.Set("description", r.Description)
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("tunnel_id", r.TunnelId)
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxVpnTunnelUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := vpnTunnelRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, vpnTunnelAPIPath, data, resourceNetboxVpnTunnelRead)
}

func resourceNetboxVpnTunnelDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, vpnTunnelAPIPath)
}
