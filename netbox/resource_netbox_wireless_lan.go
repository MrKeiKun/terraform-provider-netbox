package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxWirelessLANStatusOptions = []string{"active", "reserved", "disabled", "deprecated"}
var resourceNetboxWirelessLANAuthTypeOptions = []string{"open", "wep", "wpa-personal", "wpa-enterprise"}
var resourceNetboxWirelessLANAuthCipherOptions = []string{"auto", "tkip", "aes"}

func resourceNetboxWirelessLAN() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxWirelessLANCreate,
		Read:   resourceNetboxWirelessLANRead,
		Update: resourceNetboxWirelessLANUpdate,
		Delete: resourceNetboxWirelessLANDelete,

		Description: `:meta:subcategory:Wireless:

> A Wireless LAN represents a broadcast wireless network, identified by its SSID and optional authentication settings.`,

		Schema: map[string]*schema.Schema{
			"ssid": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice(resourceNetboxWirelessLANStatusOptions, false),
				Description:  buildValidValueDescription(resourceNetboxWirelessLANStatusOptions),
			},
			"group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"vlan_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"auth_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxWirelessLANAuthTypeOptions, false),
				Description:  buildValidValueDescription(resourceNetboxWirelessLANAuthTypeOptions),
			},
			"auth_cipher": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxWirelessLANAuthCipherOptions, false),
				Description:  buildValidValueDescription(resourceNetboxWirelessLANAuthCipherOptions),
			},
			"auth_psk": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"scope_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scope_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const wirelessLANAPIPath = "wireless/wireless-lans/"

type wirelessLANResponse struct {
	ScopeID      *int64                 `json:"scope_id"`
	ScopeType    string                 `json:"scope_type"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Ssid         string                 `json:"ssid"`
	Description  string                 `json:"description"`
	Comments     string                 `json:"comments"`
	AuthPsk      string                 `json:"auth_psk"`
	Status       *netboxapi.Choice      `json:"status"`
	AuthType     *netboxapi.Choice      `json:"auth_type"`
	AuthCipher   *netboxapi.Choice      `json:"auth_cipher"`
	Group        *netboxapi.Ref         `json:"group"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	Vlan         *netboxapi.Ref         `json:"vlan"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

// wirelessLANBody builds the update. Values that are not set are sent as null,
// so that they are cleared in NetBox instead of being left as they were.
func wirelessLANBody(api *providerState, d *schema.ResourceData) (map[string]interface{}, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"ssid":        d.Get("ssid").(string),
		"status":      d.Get("status").(string),
		"description": getOptionalStr(d, "description", true),
		"comments":    getOptionalStr(d, "comments", true),
		"owner":       getOptionalInt(d, "owner_id"),
		"group":       nullableID(d.Get("group_id").(int)),
		"tenant":      nullableID(d.Get("tenant_id").(int)),
		"vlan":        nullableID(d.Get("vlan_id").(int)),
		"tags":        tags,
	}
	setBodyStr(body, d, "scope_type", "scope_type")
	body["scope_id"] = getOptionalInt(d, "scope_id")

	if authType, ok := d.GetOk("auth_type"); ok {
		body["auth_type"] = authType.(string)
	} else {
		body["auth_type"] = nil
	}
	if authCipher, ok := d.GetOk("auth_cipher"); ok {
		body["auth_cipher"] = authCipher.(string)
	} else {
		body["auth_cipher"] = nil
	}
	if authPSK, ok := d.GetOk("auth_psk"); ok {
		body["auth_psk"] = authPSK.(string)
	} else {
		body["auth_psk"] = ""
	}

	if cf, ok := managedCustomFields(d); ok {
		body["custom_fields"] = cf
	}
	return body, nil
}

func resourceNetboxWirelessLANCreate(d *schema.ResourceData, m interface{}) error {
	body, err := wirelessLANBody(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, wirelessLANAPIPath, body, resourceNetboxWirelessLANRead)
}

func resourceNetboxWirelessLANRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, wirelessLANAPIPath, func(wlan *wirelessLANResponse) {
		d.Set("scope_id", wlan.ScopeID)
		d.Set("scope_type", wlan.ScopeType)
		d.Set("ssid", wlan.Ssid)
		d.Set("description", wlan.Description)
		d.Set("comments", wlan.Comments)
		d.Set("auth_psk", wlan.AuthPsk)
		d.Set("status", netboxapi.ChoiceValue(wlan.Status))
		d.Set("auth_type", netboxapi.ChoiceValue(wlan.AuthType))
		d.Set("auth_cipher", netboxapi.ChoiceValue(wlan.AuthCipher))
		d.Set("group_id", netboxapi.RefID(wlan.Group))
		d.Set("tenant_id", netboxapi.RefID(wlan.Tenant))
		d.Set("vlan_id", netboxapi.RefID(wlan.Vlan))
		d.Set(customFieldsKey, getCustomFields(wlan.CustomFields))
		api.readTagRefs(d, wlan.Tags)
		d.Set("owner_id", netboxapi.RefID(wlan.Owner))
	})
}

func resourceNetboxWirelessLANUpdate(d *schema.ResourceData, m interface{}) error {
	body, err := wirelessLANBody(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, wirelessLANAPIPath, body, resourceNetboxWirelessLANRead)
}

func resourceNetboxWirelessLANDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, wirelessLANAPIPath)
}
