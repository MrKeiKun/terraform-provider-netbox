package netbox

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxInterfaceModeOptions = []string{"access", "tagged", "tagged-all"}

func resourceNetboxInterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxInterfaceCreate,
		ReadContext:   resourceNetboxInterfaceRead,
		UpdateContext: resourceNetboxInterfaceUpdate,
		DeleteContext: resourceNetboxInterfaceDelete,

		Description: `:meta:subcategory:Virtualization:From the [official documentation](https://docs.netbox.dev/en/stable/features/virtualization/#interfaces):

> Virtual machine interfaces behave similarly to device interfaces, and can be assigned to VRFs, and may have IP addresses, VLANs, and services attached to them. However, given their virtual nature, they lack properties pertaining to physical attributes. For example, VM interfaces do not have a physical type and cannot have cables attached to them.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"virtual_machine_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"primary_mac_address_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The primary MAC address id.",
			},
			"mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxInterfaceModeOptions, false),
				Description:  buildValidValueDescription(resourceNetboxInterfaceModeOptions),
			},
			"mtu": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 65536),
			},
			"type": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "This attribute is not supported by netbox any longer. It will be removed in future versions of this provider.",
			},
			tagsKey: tagsSchema,
			"tagged_vlans": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"untagged_vlan": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"bridge_interface_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ID of the bridge interface this interface belongs to",
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"parent": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"qinq_svlan_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"vlan_translation_policy_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"vrf_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const vmInterfaceAPIPath = "virtualization/interfaces/"

type vmInterfaceResponse struct {
	Vrf                   *netboxapi.Ref     `json:"vrf"`
	VlanTranslationPolicy *netboxapi.Ref     `json:"vlan_translation_policy"`
	QinqSvlan             *netboxapi.Ref     `json:"qinq_svlan"`
	Parent                *int64             `json:"parent"`
	Name                  string             `json:"name"`
	Description           string             `json:"description"`
	Enabled               bool               `json:"enabled"`
	MacAddress            *string            `json:"mac_address"`
	Mtu                   *int64             `json:"mtu"`
	Tags                  []netboxapi.TagRef `json:"tags"`
	TaggedVlans           []netboxapi.Ref    `json:"tagged_vlans"`
	VirtualMachine        *netboxapi.Ref     `json:"virtual_machine"`
	Mode                  *netboxapi.Choice  `json:"mode"`
	UntaggedVlan          *netboxapi.Ref     `json:"untagged_vlan"`
	Bridge                *netboxapi.Ref     `json:"bridge"`
	PrimaryMacAddress     *netboxapi.Ref     `json:"primary_mac_address"`
	Owner                 *netboxapi.Ref     `json:"owner"`
}

// vmInterfaceBody builds the fields that are sent on both create and update.
func vmInterfaceBody(api *providerState, d *schema.ResourceData) (map[string]interface{}, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"name":            d.Get("name").(string),
		"description":     d.Get("description").(string),
		"enabled":         d.Get("enabled").(bool),
		"tags":            tags,
		"tagged_vlans":    toInt64List(d.Get("tagged_vlans")),
		"virtual_machine": int64(d.Get("virtual_machine_id").(int)),
		"owner":           getOptionalInt(d, "owner_id"),
	}
	body["parent"] = getOptionalInt(d, "parent")
	body["qinq_svlan"] = getOptionalInt(d, "qinq_svlan_id")
	body["vlan_translation_policy"] = getOptionalInt(d, "vlan_translation_policy_id")
	body["vrf"] = getOptionalInt(d, "vrf_id")
	if mode := d.Get("mode").(string); mode != "" {
		body["mode"] = mode
	}
	return body, nil
}

func resourceNetboxInterfaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)

	body, err := vmInterfaceBody(api, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if mtu, ok := d.Get("mtu").(int); ok && mtu != 0 {
		body["mtu"] = int64(mtu)
	}
	if untaggedVlan, ok := d.Get("untagged_vlan").(int); ok && untaggedVlan != 0 {
		body["untagged_vlan"] = int64(untaggedVlan)
	}
	if bridgeIF, ok := d.Get("bridge_interface_id").(int); ok && bridgeIF != 0 {
		body["bridge"] = int64(bridgeIF)
	}

	return createObjectCtx(ctx, d, m, vmInterfaceAPIPath, body, resourceNetboxInterfaceRead)
}

func resourceNetboxInterfaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)
	return readObjectCtx(ctx, d, m, vmInterfaceAPIPath, func(iface *vmInterfaceResponse) {
		d.Set("vrf_id", netboxapi.RefID(iface.Vrf))
		d.Set("vlan_translation_policy_id", netboxapi.RefID(iface.VlanTranslationPolicy))
		d.Set("qinq_svlan_id", netboxapi.RefID(iface.QinqSvlan))
		d.Set("parent", iface.Parent)
		d.Set("name", iface.Name)
		d.Set("description", iface.Description)
		d.Set("enabled", iface.Enabled)
		d.Set("mac_address", iface.MacAddress)
		d.Set("mtu", iface.Mtu)
		api.readTagRefs(d, iface.Tags)
		d.Set("owner_id", netboxapi.RefID(iface.Owner))
		d.Set("tagged_vlans", netboxapi.IDs(iface.TaggedVlans))
		if iface.VirtualMachine != nil {
			d.Set("virtual_machine_id", iface.VirtualMachine.ID)
		}
		if iface.Mode != nil {
			d.Set("mode", iface.Mode.Value)
		}
		if iface.UntaggedVlan != nil {
			d.Set("untagged_vlan", iface.UntaggedVlan.ID)
		}
		if iface.Bridge != nil {
			d.Set("bridge_interface_id", iface.Bridge.ID)
		} else {
			d.Set("bridge_interface_id", nil)
		}
		if iface.PrimaryMacAddress != nil {
			d.Set("primary_mac_address_id", iface.PrimaryMacAddress.ID)
		}
	})
}

func resourceNetboxInterfaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)

	body, err := vmInterfaceBody(api, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChange("mac_address") {
		body["mac_address"] = d.Get("mac_address").(string)
	}
	if d.HasChange("mtu") {
		body["mtu"] = nullableID(d.Get("mtu").(int))
	}
	if d.HasChange("untagged_vlan") {
		body["untagged_vlan"] = nullableID(d.Get("untagged_vlan").(int))
	}
	if d.HasChange("primary_mac_address_id") {
		body["primary_mac_address"] = nullableID(d.Get("primary_mac_address_id").(int))
	}
	// A bridge that is no longer set has to be sent as an explicit null,
	// otherwise NetBox keeps the old one.
	if d.HasChange("bridge_interface_id") {
		body["bridge"] = nullableID(d.Get("bridge_interface_id").(int))
	}

	return patchObjectCtx(ctx, d, m, vmInterfaceAPIPath, body, resourceNetboxInterfaceRead)
}

func resourceNetboxInterfaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, vmInterfaceAPIPath)
}
