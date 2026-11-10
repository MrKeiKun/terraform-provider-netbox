package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxVpnTunnelGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxVpnTunnelGroupCreate),
		ReadContext:   withContext(resourceNetboxVpnTunnelGroupRead),
		UpdateContext: withContext(resourceNetboxVpnTunnelGroupUpdate),
		DeleteContext: withContext(resourceNetboxVpnTunnelGroupDelete),

		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/features/vpn-tunnels/):

> NetBox can model private tunnels formed among virtual termination points across your network. Typical tunnel implementations include GRE, IP-in-IP, and IPSec. A tunnel may be terminated to two or more device or virtual machine interfaces. For convenient organization, tunnels may be assigned to user-defined groups.`,

		Schema: map[string]*schema.Schema{
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
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the vpn tunnel group.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the vpn tunnel group.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const vpnTunnelGroupAPIPath = "vpn/tunnel-groups/"

type vpnTunnelGroupRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type vpnTunnelGroupResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func vpnTunnelGroupRequestFromResourceData(d *schema.ResourceData) *vpnTunnelGroupRequest {
	name := d.Get("name").(string)
	data := &vpnTunnelGroupRequest{
		CustomFields: customFieldsRequest(d),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Description:  getOptionalStr(d, "description", true),
		Tags:         []netboxapi.TagRef{},
	}
	return data
}

func resourceNetboxVpnTunnelGroupCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, vpnTunnelGroupAPIPath, vpnTunnelGroupRequestFromResourceData(d), resourceNetboxVpnTunnelGroupRead)
}

func resourceNetboxVpnTunnelGroupRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, vpnTunnelGroupAPIPath, func(r *vpnTunnelGroupResponse) {
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
	})
}

func resourceNetboxVpnTunnelGroupUpdate(d *schema.ResourceData, m interface{}) error {
	return putObject(d, m, vpnTunnelGroupAPIPath, vpnTunnelGroupRequestFromResourceData(d), resourceNetboxVpnTunnelGroupRead)
}

func resourceNetboxVpnTunnelGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, vpnTunnelGroupAPIPath)
}
