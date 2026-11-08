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
				Type:        schema.TypeString,
				Description: "The name of the vpn tunnel.",
				Required:    true,
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
				Type:        schema.TypeInt,
				Description: "The ID of the tunnel group associated with the vpn tunnel.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the vpn tunnel.",
				Optional:    true,
			},
			"tunnel_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tunnel associated with the vpn tunnel.",
				Optional:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the vpn tunnel.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the vpn tunnel.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the vpn tunnel.",
				Optional:    true,
			},
			"ipsec_profile_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the ipsec profile associated with the vpn tunnel.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
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
	Owner         *int64             `json:"owner"`
	Comments      string             `json:"comments,omitempty"`
	IpsecProfile  *int64             `json:"ipsec_profile"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

type vpnTunnelResponse struct {
	Name          string                 `json:"name"`
	Encapsulation *netboxapi.Choice      `json:"encapsulation"`
	Status        *netboxapi.Choice      `json:"status"`
	Group         *netboxapi.Ref         `json:"group"`
	Description   string                 `json:"description"`
	Tenant        *netboxapi.Ref         `json:"tenant"`
	TunnelId      *int64                 `json:"tunnel_id"`
	Owner         *netboxapi.Ref         `json:"owner"`
	Comments      string                 `json:"comments"`
	IpsecProfile  *netboxapi.Ref         `json:"ipsec_profile"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
	Tags          []netboxapi.TagRef     `json:"tags"`
}

func vpnTunnelRequestFromResourceData(api *providerState, d *schema.ResourceData) (*vpnTunnelRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &vpnTunnelRequest{
		CustomFields:  customFieldsRequest(d),
		IpsecProfile:  getOptionalInt(d, "ipsec_profile_id"),
		Comments:      getOptionalStr(d, "comments", true),
		Owner:         getOptionalInt(d, "owner_id"),
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
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("ipsec_profile_id", netboxapi.RefID(r.IpsecProfile))
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
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
