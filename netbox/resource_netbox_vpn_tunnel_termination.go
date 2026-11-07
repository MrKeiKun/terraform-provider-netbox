package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var resourceNetboxVpnTunnelTerminationRoleOptions = []string{"peer", "hub", "spoke"}

func resourceNetboxVpnTunnelTermination() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVpnTunnelTerminationCreate,
		Read:   resourceNetboxVpnTunnelTerminationRead,
		Update: resourceNetboxVpnTunnelTerminationUpdate,
		Delete: resourceNetboxVpnTunnelTerminationDelete,

		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/features/vpn-tunnels/):

> NetBox can model private tunnels formed among virtual termination points across your network. Typical tunnel implementations include GRE, IP-in-IP, and IPSec. A tunnel may be terminated to two or more device or virtual machine interfaces. For convenient organization, tunnels may be assigned to user-defined groups.`,

		Schema: map[string]*schema.Schema{
			"tunnel_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				Description: buildValidValueDescription(resourceNetboxVpnTunnelTerminationRoleOptions),
			},
			"virtual_machine_interface_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ExactlyOneOf: []string{"virtual_machine_interface_id", "device_interface_id"},
			},
			"device_interface_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ExactlyOneOf: []string{"virtual_machine_interface_id", "device_interface_id"},
			},
			"outside_ip_address_id": {
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

const vpnTunnelTerminationAPIPath = "vpn/tunnel-terminations/"

type vpnTunnelTerminationRequest struct {
	Tunnel          int64              `json:"tunnel"`
	Role            string             `json:"role,omitempty"`
	TerminationType *string            `json:"termination_type"`
	TerminationID   int64              `json:"termination_id,omitempty"`
	OutsideIP       *int64             `json:"outside_ip,omitempty"`
	Tags            []netboxapi.TagRef `json:"tags"`
}

type vpnTunnelTerminationResponse struct {
	Tunnel        *netboxapi.Ref     `json:"tunnel"`
	Role          *netboxapi.Choice  `json:"role"`
	TerminationID int64              `json:"termination_id"`
	OutsideIP     *netboxapi.Ref     `json:"outside_ip"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

func vpnTunnelTerminationRequestFromResourceData(api *providerState, d *schema.ResourceData) *vpnTunnelTerminationRequest {
	// A failed tag lookup leaves the tags empty, same as before.
	tags, _ := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))

	data := &vpnTunnelTerminationRequest{
		Tunnel:    int64(d.Get("tunnel_id").(int)),
		Role:      d.Get("role").(string),
		OutsideIP: getOptionalInt(d, "outside_ip_address_id"),
		Tags:      tags,
	}

	vmInterfaceID := getOptionalInt(d, "virtual_machine_interface_id")
	deviceInterfaceID := getOptionalInt(d, "device_interface_id")
	switch {
	case vmInterfaceID != nil:
		data.TerminationType = strToPtr("virtualization.vminterface")
		data.TerminationID = *vmInterfaceID
	case deviceInterfaceID != nil:
		data.TerminationType = strToPtr("dcim.interface")
		data.TerminationID = *deviceInterfaceID
	}
	return data
}

func resourceNetboxVpnTunnelTerminationCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, vpnTunnelTerminationAPIPath, vpnTunnelTerminationRequestFromResourceData(m.(*providerState), d), resourceNetboxVpnTunnelTerminationRead)
}

func resourceNetboxVpnTunnelTerminationRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, vpnTunnelTerminationAPIPath, func(tunnelTermination *vpnTunnelTerminationResponse) {
		if tunnelTermination.Tunnel != nil {
			d.Set("tunnel_id", tunnelTermination.Tunnel.ID)
		}
		d.Set("role", netboxapi.ChoiceValue(tunnelTermination.Role))

		vmInterfaceID := getOptionalInt(d, "virtual_machine_interface_id")
		deviceInterfaceID := getOptionalInt(d, "device_interface_id")
		switch {
		case vmInterfaceID != nil:
			d.Set("virtual_machine_interface_id", tunnelTermination.TerminationID)
		case deviceInterfaceID != nil:
			d.Set("device_interface_id", tunnelTermination.TerminationID)
		}

		if tunnelTermination.OutsideIP != nil {
			d.Set("outside_ip_address_id", tunnelTermination.OutsideIP.ID)
		}
		api.readTagRefs(d, tunnelTermination.Tags)
	})
}

func resourceNetboxVpnTunnelTerminationUpdate(d *schema.ResourceData, m interface{}) error {
	return putObject(d, m, vpnTunnelTerminationAPIPath, vpnTunnelTerminationRequestFromResourceData(m.(*providerState), d), resourceNetboxVpnTunnelTerminationRead)
}

func resourceNetboxVpnTunnelTerminationDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, vpnTunnelTerminationAPIPath)
}
