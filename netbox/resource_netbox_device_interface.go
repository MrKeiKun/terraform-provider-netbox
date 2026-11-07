package netbox

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxDeviceInterfaceModeOptions = []string{"access", "tagged", "tagged-all", "q-in-q"}

func resourceNetboxDeviceInterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxDeviceInterfaceCreate,
		ReadContext:   resourceNetboxDeviceInterfaceRead,
		UpdateContext: resourceNetboxDeviceInterfaceUpdate,
		DeleteContext: resourceNetboxDeviceInterfaceDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/device/#interface):

> Interfaces in NetBox represent network interfaces used to exchange data with connected devices. On modern networks, these are most commonly Ethernet, but other types are supported as well. IP addresses and VLANs can be assigned to interfaces.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"device_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"module_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"lag_device_interface_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "If this device is a member of a LAG group, you can reference the LAG interface here.",
			},
			"mac_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The MAC address as string from the first MAC address assigned to this interface, if any.",
			},
			"mac_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"primary_mac_address_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The primary MAC address id.",
			},
			"mgmtonly": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxDeviceInterfaceModeOptions, false),
				Description:  buildValidValueDescription(resourceNetboxDeviceInterfaceModeOptions),
			},
			"mtu": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 65536),
			},
			"parent_device_interface_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The netbox_device_interface id of the parent interface. Useful if this interface is a logical interface.",
			},
			"speed": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
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

const deviceInterfaceAPIPath = "dcim/interfaces/"

type deviceInterfaceMAC struct {
	ID          int64  `json:"id"`
	MacAddress  string `json:"mac_address"`
	Description string `json:"description"`
}

type deviceInterfaceResponse struct {
	Name              string               `json:"name"`
	Description       string               `json:"description"`
	Label             string               `json:"label"`
	Type              *netboxapi.Choice    `json:"type"`
	Enabled           bool                 `json:"enabled"`
	MgmtOnly          bool                 `json:"mgmt_only"`
	Mtu               *int64               `json:"mtu"`
	Speed             *int64               `json:"speed"`
	Tags              []netboxapi.TagRef   `json:"tags"`
	TaggedVlans       []netboxapi.Ref      `json:"tagged_vlans"`
	Device            *netboxapi.Ref       `json:"device"`
	Module            *netboxapi.Ref       `json:"module"`
	Lag               *netboxapi.Ref       `json:"lag"`
	Mode              *netboxapi.Choice    `json:"mode"`
	Parent            *netboxapi.Ref       `json:"parent"`
	UntaggedVlan      *netboxapi.Ref       `json:"untagged_vlan"`
	Vrf               *netboxapi.Ref       `json:"vrf"`
	MacAddresses      []deviceInterfaceMAC `json:"mac_addresses"`
	PrimaryMacAddress *netboxapi.Ref       `json:"primary_mac_address"`
}

// deviceInterfaceBody builds the fields that are sent on both create and update.
func deviceInterfaceBody(api *providerState, d *schema.ResourceData) (map[string]interface{}, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"name":          d.Get("name").(string),
		"description":   d.Get("description").(string),
		"label":         d.Get("label").(string),
		"type":          d.Get("type").(string),
		"enabled":       d.Get("enabled").(bool),
		"mgmt_only":     d.Get("mgmtonly").(bool),
		"tags":          tags,
		"tagged_vlans":  toInt64List(d.Get("tagged_vlans")),
		"device":        int64(d.Get("device_id").(int)),
		"wireless_lans": []int64{},
		"vdcs":          []int64{},
	}
	if mode := d.Get("mode").(string); mode != "" {
		body["mode"] = mode
	}
	return body, nil
}

func resourceNetboxDeviceInterfaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)

	body, err := deviceInterfaceBody(api, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if lag, ok := d.Get("lag_device_interface_id").(int); ok && lag != 0 {
		body["lag"] = int64(lag)
	}
	if mtu, ok := d.Get("mtu").(int); ok && mtu != 0 {
		body["mtu"] = int64(mtu)
	}
	if parent, ok := d.Get("parent_device_interface_id").(int); ok && parent != 0 {
		body["parent"] = int64(parent)
	}
	if speed, ok := d.Get("speed").(int); ok && speed != 0 {
		body["speed"] = int64(speed)
	}
	if untaggedVlan, ok := d.Get("untagged_vlan").(int); ok && untaggedVlan != 0 {
		body["untagged_vlan"] = int64(untaggedVlan)
	}
	if vrf := getOptionalInt(d, "vrf_id"); vrf != nil {
		body["vrf"] = *vrf
	}
	if module := getOptionalInt(d, "module_id"); module != nil {
		body["module"] = *module
	}

	return createObjectCtx(ctx, d, m, deviceInterfaceAPIPath, body, resourceNetboxDeviceInterfaceRead)
}

func resourceNetboxDeviceInterfaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)
	return readObjectCtx(ctx, d, m, deviceInterfaceAPIPath, func(iface *deviceInterfaceResponse) {
		d.Set("name", iface.Name)
		d.Set("description", iface.Description)
		d.Set("label", iface.Label)
		d.Set("type", netboxapi.ChoiceValue(iface.Type))
		d.Set("enabled", iface.Enabled)
		d.Set("mgmtonly", iface.MgmtOnly)
		d.Set("mtu", iface.Mtu)
		d.Set("speed", iface.Speed)
		api.readTagRefs(d, iface.Tags)
		d.Set("tagged_vlans", netboxapi.IDs(iface.TaggedVlans))
		if iface.Device != nil {
			d.Set("device_id", iface.Device.ID)
		}
		d.Set("module_id", netboxapi.RefID(iface.Module))
		if iface.Lag != nil {
			d.Set("lag_device_interface_id", iface.Lag.ID)
		}
		if iface.Mode != nil {
			d.Set("mode", iface.Mode.Value)
		}
		if iface.Parent != nil {
			d.Set("parent_device_interface_id", iface.Parent.ID)
		}
		if iface.UntaggedVlan != nil {
			d.Set("untagged_vlan", iface.UntaggedVlan.ID)
		}
		d.Set("vrf_id", netboxapi.RefID(iface.Vrf))
		if iface.MacAddresses != nil {
			var macAddresses []map[string]interface{}
			for i, mac := range iface.MacAddresses {
				// We just set the first mac address in the `mac_address` attribute
				if i == 0 {
					d.Set("mac_address", mac.MacAddress)
				}
				macAddresses = append(macAddresses, map[string]interface{}{
					"id":          mac.ID,
					"description": mac.Description,
					"mac_address": mac.MacAddress,
				})
			}
			d.Set("mac_addresses", macAddresses)
		}
		if iface.PrimaryMacAddress != nil {
			d.Set("primary_mac_address_id", iface.PrimaryMacAddress.ID)
		}
	})
}

func resourceNetboxDeviceInterfaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)

	body, err := deviceInterfaceBody(api, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if module := getOptionalInt(d, "module_id"); module != nil {
		body["module"] = *module
	}
	if d.HasChange("lag_device_interface_id") {
		body["lag"] = nullableID(d.Get("lag_device_interface_id").(int))
	}
	if d.HasChange("mtu") {
		body["mtu"] = nullableID(d.Get("mtu").(int))
	}
	if d.HasChange("parent_device_interface_id") {
		body["parent"] = nullableID(d.Get("parent_device_interface_id").(int))
	}
	if d.HasChange("speed") {
		body["speed"] = nullableID(d.Get("speed").(int))
	}
	if d.HasChange("untagged_vlan") {
		body["untagged_vlan"] = nullableID(d.Get("untagged_vlan").(int))
	}
	if d.HasChange("primary_mac_address_id") {
		body["primary_mac_address"] = nullableID(d.Get("primary_mac_address_id").(int))
	}
	if d.HasChange("vrf_id") {
		body["vrf"] = nullableID(d.Get("vrf_id").(int))
	}

	return patchObjectCtx(ctx, d, m, deviceInterfaceAPIPath, body, resourceNetboxDeviceInterfaceRead)
}

func resourceNetboxDeviceInterfaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, deviceInterfaceAPIPath)
}
