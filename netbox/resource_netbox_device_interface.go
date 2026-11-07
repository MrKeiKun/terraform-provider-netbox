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
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"channel_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 1024),
				Description:  "The channel on the parent interface to which this subinterface is bound",
			},
			"bridge": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"duplex": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [half, full, auto]",
				ValidateFunc: validation.StringInSlice([]string{"half", "full", "auto"}, false),
			},
			"wwn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rf_role": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [ap, station]",
				ValidateFunc: validation.StringInSlice([]string{"ap", "station"}, false),
			},
			"rf_channel": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [2.4g-1-2412-22, 2.4g-2-2417-22, 2.4g-3-2422-22, 2.4g-4-2427-22, 2.4g-5-2432-22, 2.4g-6-2437-22, 2.4g-7-2442-22, 2.4g-8-2447-22, 2.4g-9-2452-22, 2.4g-10-2457-22, 2.4g-11-2462-22, 2.4g-12-2467-22, 2.4g-13-2472-22, 5g-32-5160-20, 5g-34-5170-40, 5g-36-5180-20, 5g-38-5190-40, 5g-40-5200-20, 5g-42-5210-80, 5g-44-5220-20, 5g-46-5230-40, 5g-48-5240-20, 5g-50-5250-160, 5g-52-5260-20, 5g-54-5270-40, 5g-56-5280-20, 5g-58-5290-80, 5g-60-5300-20, 5g-62-5310-40, 5g-64-5320-20, 5g-100-5500-20, 5g-102-5510-40, 5g-104-5520-20, 5g-106-5530-80, 5g-108-5540-20, 5g-110-5550-40, 5g-112-5560-20, 5g-114-5570-160, 5g-116-5580-20, 5g-118-5590-40, 5g-120-5600-20, 5g-122-5610-80, 5g-124-5620-20, 5g-126-5630-40, 5g-128-5640-20, 5g-132-5660-20, 5g-134-5670-40, 5g-136-5680-20, 5g-138-5690-80, 5g-140-5700-20, 5g-142-5710-40, 5g-144-5720-20, 5g-149-5745-20, 5g-151-5755-40, 5g-153-5765-20, 5g-155-5775-80, 5g-157-5785-20, 5g-159-5795-40, 5g-161-5805-20, 5g-163-5815-160, 5g-165-5825-20, 5g-167-5835-40, 5g-169-5845-20, 5g-171-5855-80, 5g-173-5865-20, 5g-175-5875-40, 5g-177-5885-20, 6g-1-5955-20, 6g-3-5965-40, 6g-5-5975-20, 6g-7-5985-80, 6g-9-5995-20, 6g-11-6005-40, 6g-13-6015-20, 6g-15-6025-160, 6g-17-6035-20, 6g-19-6045-40, 6g-21-6055-20, 6g-23-6065-80, 6g-25-6075-20, 6g-27-6085-40, 6g-29-6095-20, 6g-31-6105-320, 6g-33-6115-20, 6g-35-6125-40, 6g-37-6135-20, 6g-39-6145-80, 6g-41-6155-20, 6g-43-6165-40, 6g-45-6175-20, 6g-47-6185-160, 6g-49-6195-20, 6g-51-6205-40, 6g-53-6215-20, 6g-55-6225-80, 6g-57-6235-20, 6g-59-6245-40, 6g-61-6255-20, 6g-65-6275-20, 6g-67-6285-40, 6g-69-6295-20, 6g-71-6305-80, 6g-73-6315-20, 6g-75-6325-40, 6g-77-6335-20, 6g-79-6345-160, 6g-81-6355-20, 6g-83-6365-40, 6g-85-6375-20, 6g-87-6385-80, 6g-89-6395-20, 6g-91-6405-40, 6g-93-6415-20, 6g-95-6425-320, 6g-97-6435-20, 6g-99-6445-40, 6g-101-6455-20, 6g-103-6465-80, 6g-105-6475-20, 6g-107-6485-40, 6g-109-6495-20, 6g-111-6505-160, 6g-113-6515-20, 6g-115-6525-40, 6g-117-6535-20, 6g-119-6545-80, 6g-121-6555-20, 6g-123-6565-40, 6g-125-6575-20, 6g-129-6595-20, 6g-131-6605-40, 6g-133-6615-20, 6g-135-6625-80, 6g-137-6635-20, 6g-139-6645-40, 6g-141-6655-20, 6g-143-6665-160, 6g-145-6675-20, 6g-147-6685-40, 6g-149-6695-20, 6g-151-6705-80, 6g-153-6715-20, 6g-155-6725-40, 6g-157-6735-20, 6g-159-6745-320, 6g-161-6755-20, 6g-163-6765-40, 6g-165-6775-20, 6g-167-6785-80, 6g-169-6795-20, 6g-171-6805-40, 6g-173-6815-20, 6g-175-6825-160, 6g-177-6835-20, 6g-179-6845-40, 6g-181-6855-20, 6g-183-6865-80, 6g-185-6875-20, 6g-187-6885-40, 6g-189-6895-20, 6g-193-6915-20, 6g-195-6925-40, 6g-197-6935-20, 6g-199-6945-80, 6g-201-6955-20, 6g-203-6965-40, 6g-205-6975-20, 6g-207-6985-160, 6g-209-6995-20, 6g-211-7005-40, 6g-213-7015-20, 6g-215-7025-80, 6g-217-7035-20, 6g-219-7045-40, 6g-221-7055-20, 6g-225-7075-20, 6g-227-7085-40, 6g-229-7095-20, 6g-233-7115-20, 60g-1-58320-2160, 60g-2-60480-2160, 60g-3-62640-2160, 60g-4-64800-2160, 60g-5-66960-2160, 60g-6-69120-2160, 60g-9-59400-4320, 60g-10-61560-4320, 60g-11-63720-4320, 60g-12-65880-4320, 60g-13-68040-4320, 60g-17-60480-6480, 60g-18-62640-6480, 60g-19-64800-6480, 60g-20-66960-6480, 60g-25-61560-6480, 60g-26-63720-6480, 60g-27-65880-6480]",
				ValidateFunc: validation.StringInSlice([]string{"2.4g-1-2412-22", "2.4g-2-2417-22", "2.4g-3-2422-22", "2.4g-4-2427-22", "2.4g-5-2432-22", "2.4g-6-2437-22", "2.4g-7-2442-22", "2.4g-8-2447-22", "2.4g-9-2452-22", "2.4g-10-2457-22", "2.4g-11-2462-22", "2.4g-12-2467-22", "2.4g-13-2472-22", "5g-32-5160-20", "5g-34-5170-40", "5g-36-5180-20", "5g-38-5190-40", "5g-40-5200-20", "5g-42-5210-80", "5g-44-5220-20", "5g-46-5230-40", "5g-48-5240-20", "5g-50-5250-160", "5g-52-5260-20", "5g-54-5270-40", "5g-56-5280-20", "5g-58-5290-80", "5g-60-5300-20", "5g-62-5310-40", "5g-64-5320-20", "5g-100-5500-20", "5g-102-5510-40", "5g-104-5520-20", "5g-106-5530-80", "5g-108-5540-20", "5g-110-5550-40", "5g-112-5560-20", "5g-114-5570-160", "5g-116-5580-20", "5g-118-5590-40", "5g-120-5600-20", "5g-122-5610-80", "5g-124-5620-20", "5g-126-5630-40", "5g-128-5640-20", "5g-132-5660-20", "5g-134-5670-40", "5g-136-5680-20", "5g-138-5690-80", "5g-140-5700-20", "5g-142-5710-40", "5g-144-5720-20", "5g-149-5745-20", "5g-151-5755-40", "5g-153-5765-20", "5g-155-5775-80", "5g-157-5785-20", "5g-159-5795-40", "5g-161-5805-20", "5g-163-5815-160", "5g-165-5825-20", "5g-167-5835-40", "5g-169-5845-20", "5g-171-5855-80", "5g-173-5865-20", "5g-175-5875-40", "5g-177-5885-20", "6g-1-5955-20", "6g-3-5965-40", "6g-5-5975-20", "6g-7-5985-80", "6g-9-5995-20", "6g-11-6005-40", "6g-13-6015-20", "6g-15-6025-160", "6g-17-6035-20", "6g-19-6045-40", "6g-21-6055-20", "6g-23-6065-80", "6g-25-6075-20", "6g-27-6085-40", "6g-29-6095-20", "6g-31-6105-320", "6g-33-6115-20", "6g-35-6125-40", "6g-37-6135-20", "6g-39-6145-80", "6g-41-6155-20", "6g-43-6165-40", "6g-45-6175-20", "6g-47-6185-160", "6g-49-6195-20", "6g-51-6205-40", "6g-53-6215-20", "6g-55-6225-80", "6g-57-6235-20", "6g-59-6245-40", "6g-61-6255-20", "6g-65-6275-20", "6g-67-6285-40", "6g-69-6295-20", "6g-71-6305-80", "6g-73-6315-20", "6g-75-6325-40", "6g-77-6335-20", "6g-79-6345-160", "6g-81-6355-20", "6g-83-6365-40", "6g-85-6375-20", "6g-87-6385-80", "6g-89-6395-20", "6g-91-6405-40", "6g-93-6415-20", "6g-95-6425-320", "6g-97-6435-20", "6g-99-6445-40", "6g-101-6455-20", "6g-103-6465-80", "6g-105-6475-20", "6g-107-6485-40", "6g-109-6495-20", "6g-111-6505-160", "6g-113-6515-20", "6g-115-6525-40", "6g-117-6535-20", "6g-119-6545-80", "6g-121-6555-20", "6g-123-6565-40", "6g-125-6575-20", "6g-129-6595-20", "6g-131-6605-40", "6g-133-6615-20", "6g-135-6625-80", "6g-137-6635-20", "6g-139-6645-40", "6g-141-6655-20", "6g-143-6665-160", "6g-145-6675-20", "6g-147-6685-40", "6g-149-6695-20", "6g-151-6705-80", "6g-153-6715-20", "6g-155-6725-40", "6g-157-6735-20", "6g-159-6745-320", "6g-161-6755-20", "6g-163-6765-40", "6g-165-6775-20", "6g-167-6785-80", "6g-169-6795-20", "6g-171-6805-40", "6g-173-6815-20", "6g-175-6825-160", "6g-177-6835-20", "6g-179-6845-40", "6g-181-6855-20", "6g-183-6865-80", "6g-185-6875-20", "6g-187-6885-40", "6g-189-6895-20", "6g-193-6915-20", "6g-195-6925-40", "6g-197-6935-20", "6g-199-6945-80", "6g-201-6955-20", "6g-203-6965-40", "6g-205-6975-20", "6g-207-6985-160", "6g-209-6995-20", "6g-211-7005-40", "6g-213-7015-20", "6g-215-7025-80", "6g-217-7035-20", "6g-219-7045-40", "6g-221-7055-20", "6g-225-7075-20", "6g-227-7085-40", "6g-229-7095-20", "6g-233-7115-20", "60g-1-58320-2160", "60g-2-60480-2160", "60g-3-62640-2160", "60g-4-64800-2160", "60g-5-66960-2160", "60g-6-69120-2160", "60g-9-59400-4320", "60g-10-61560-4320", "60g-11-63720-4320", "60g-12-65880-4320", "60g-13-68040-4320", "60g-17-60480-6480", "60g-18-62640-6480", "60g-19-64800-6480", "60g-20-66960-6480", "60g-25-61560-6480", "60g-26-63720-6480", "60g-27-65880-6480"}, false),
			},
			"poe_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [pd, pse]",
				ValidateFunc: validation.StringInSlice([]string{"pd", "pse"}, false),
			},
			"poe_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [type1-ieee802.3af, type2-ieee802.3at, type3-ieee802.3bt, type4-ieee802.3bt, passive-24v-2pair, passive-24v-4pair, passive-48v-2pair, passive-48v-4pair]",
				ValidateFunc: validation.StringInSlice([]string{"type1-ieee802.3af", "type2-ieee802.3at", "type3-ieee802.3bt", "type4-ieee802.3bt", "passive-24v-2pair", "passive-24v-4pair", "passive-48v-2pair", "passive-48v-4pair"}, false),
			},
			"rf_channel_frequency": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Populated by selected channel (if set)",
			},
			"rf_channel_width": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Populated by selected channel (if set)",
			},
			"tx_power": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(-40, 127),
			},
			"qinq_svlan_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"vlan_translation_policy_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"mark_connected": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Treat as if a cable is connected",
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
	MarkConnected         bool                 `json:"mark_connected"`
	VlanTranslationPolicy *netboxapi.Ref       `json:"vlan_translation_policy"`
	QinqSvlan             *netboxapi.Ref       `json:"qinq_svlan"`
	TxPower               *int64               `json:"tx_power"`
	RfChannelWidth        *float64             `json:"rf_channel_width"`
	RfChannelFrequency    *float64             `json:"rf_channel_frequency"`
	PoeType               *netboxapi.Choice    `json:"poe_type"`
	PoeMode               *netboxapi.Choice    `json:"poe_mode"`
	RfChannel             *netboxapi.Choice    `json:"rf_channel"`
	RfRole                *netboxapi.Choice    `json:"rf_role"`
	Wwn                   string               `json:"wwn"`
	Duplex                *netboxapi.Choice    `json:"duplex"`
	Bridge                *int64               `json:"bridge"`
	ChannelID             *int64               `json:"channel_id"`
	Owner                 *netboxapi.Ref       `json:"owner"`
	Name                  string               `json:"name"`
	Description           string               `json:"description"`
	Label                 string               `json:"label"`
	Type                  *netboxapi.Choice    `json:"type"`
	Enabled               bool                 `json:"enabled"`
	MgmtOnly              bool                 `json:"mgmt_only"`
	Mtu                   *int64               `json:"mtu"`
	Speed                 *int64               `json:"speed"`
	Tags                  []netboxapi.TagRef   `json:"tags"`
	TaggedVlans           []netboxapi.Ref      `json:"tagged_vlans"`
	Device                *netboxapi.Ref       `json:"device"`
	Module                *netboxapi.Ref       `json:"module"`
	Lag                   *netboxapi.Ref       `json:"lag"`
	Mode                  *netboxapi.Choice    `json:"mode"`
	Parent                *netboxapi.Ref       `json:"parent"`
	UntaggedVlan          *netboxapi.Ref       `json:"untagged_vlan"`
	Vrf                   *netboxapi.Ref       `json:"vrf"`
	MacAddresses          []deviceInterfaceMAC `json:"mac_addresses"`
	PrimaryMacAddress     *netboxapi.Ref       `json:"primary_mac_address"`
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
		"owner":         getOptionalInt(d, "owner_id"),
		"type":          d.Get("type").(string),
		"enabled":       d.Get("enabled").(bool),
		"mgmt_only":     d.Get("mgmtonly").(bool),
		"tags":          tags,
		"tagged_vlans":  toInt64List(d.Get("tagged_vlans")),
		"device":        int64(d.Get("device_id").(int)),
		"wireless_lans": []int64{},
		"vdcs":          []int64{},
	}
	body["channel_id"] = getOptionalInt(d, "channel_id")
	body["bridge"] = getOptionalInt(d, "bridge")
	setBodyStr(body, d, "duplex", "duplex")
	setBodyStr(body, d, "wwn", "wwn")
	setBodyStr(body, d, "rf_role", "rf_role")
	setBodyStr(body, d, "rf_channel", "rf_channel")
	setBodyStr(body, d, "poe_mode", "poe_mode")
	setBodyStr(body, d, "poe_type", "poe_type")
	body["rf_channel_frequency"] = getOptionalFloat(d, "rf_channel_frequency")
	body["rf_channel_width"] = getOptionalFloat(d, "rf_channel_width")
	body["tx_power"] = getOptionalInt(d, "tx_power")
	body["qinq_svlan"] = getOptionalInt(d, "qinq_svlan_id")
	body["vlan_translation_policy"] = getOptionalInt(d, "vlan_translation_policy_id")
	if v := getOptionalBool(d, "mark_connected"); v != nil {
		body["mark_connected"] = *v
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
		d.Set("mark_connected", iface.MarkConnected)
		d.Set("vlan_translation_policy_id", netboxapi.RefID(iface.VlanTranslationPolicy))
		d.Set("qinq_svlan_id", netboxapi.RefID(iface.QinqSvlan))
		d.Set("tx_power", iface.TxPower)
		d.Set("rf_channel_width", iface.RfChannelWidth)
		d.Set("rf_channel_frequency", iface.RfChannelFrequency)
		d.Set("poe_type", netboxapi.ChoiceValue(iface.PoeType))
		d.Set("poe_mode", netboxapi.ChoiceValue(iface.PoeMode))
		d.Set("rf_channel", netboxapi.ChoiceValue(iface.RfChannel))
		d.Set("rf_role", netboxapi.ChoiceValue(iface.RfRole))
		d.Set("wwn", iface.Wwn)
		d.Set("duplex", netboxapi.ChoiceValue(iface.Duplex))
		d.Set("bridge", iface.Bridge)
		d.Set("channel_id", iface.ChannelID)
		d.Set("name", iface.Name)
		d.Set("description", iface.Description)
		d.Set("label", iface.Label)
		d.Set("type", netboxapi.ChoiceValue(iface.Type))
		d.Set("enabled", iface.Enabled)
		d.Set("mgmtonly", iface.MgmtOnly)
		d.Set("mtu", iface.Mtu)
		d.Set("speed", iface.Speed)
		api.readTagRefs(d, iface.Tags)
		d.Set("owner_id", netboxapi.RefID(iface.Owner))
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
