package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxWirelessLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxWirelessLinkCreate),
		ReadContext:   withContext(resourceNetboxWirelessLinkRead),
		UpdateContext: withContext(resourceNetboxWirelessLinkUpdate),
		DeleteContext: withContext(resourceNetboxWirelessLinkDelete),

		Description: `:meta:subcategory:Wireless:From the [official documentation](https://docs.netbox.dev/en/stable/models/wireless/wirelesslink/):

> A wireless link connects two wireless interfaces.`,

		Schema: map[string]*schema.Schema{
			"interface_a_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the interface a associated with the wireless link.",
				Required:    true,
			},
			"interface_b_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the interface b associated with the wireless link.",
				Required:    true,
			},
			"ssid": {
				Type:        schema.TypeString,
				Description: "The ssid of the wireless link.",
				Optional:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "One of [connected, planned, decommissioning]",
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the wireless link.",
				Optional:    true,
			},
			"auth_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [open, wep, wpa-personal, wpa-enterprise]",
				ValidateFunc: validation.StringInSlice([]string{"open", "wep", "wpa-personal", "wpa-enterprise"}, false),
			},
			"auth_cipher": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [auto, tkip, aes]",
				ValidateFunc: validation.StringInSlice([]string{"auto", "tkip", "aes"}, false),
			},
			"auth_psk": {
				Type:        schema.TypeString,
				Description: "The auth psk of the wireless link.",
				Optional:    true,
			},
			"distance": {
				Type:        schema.TypeFloat,
				Description: "The distance of the wireless link.",
				Optional:    true,
			},
			"distance_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [km, m, mi, ft]",
				ValidateFunc: validation.StringInSlice([]string{"km", "m", "mi", "ft"}, false),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the wireless link.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the wireless link.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the wireless link.",
				Optional:    true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const wirelessLinkAPIPath = "wireless/wireless-links/"

type wirelessLinkRequest struct {
	InterfaceA   *int64             `json:"interface_a"`
	InterfaceB   *int64             `json:"interface_b"`
	Ssid         string             `json:"ssid,omitempty"`
	Status       string             `json:"status,omitempty"`
	Tenant       *int64             `json:"tenant"`
	AuthType     *string            `json:"auth_type,omitempty"`
	AuthCipher   *string            `json:"auth_cipher,omitempty"`
	AuthPsk      string             `json:"auth_psk,omitempty"`
	Distance     *float64           `json:"distance"`
	DistanceUnit *string            `json:"distance_unit,omitempty"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type wirelessLinkResponse struct {
	InterfaceA   *netboxapi.Ref         `json:"interface_a"`
	InterfaceB   *netboxapi.Ref         `json:"interface_b"`
	Ssid         string                 `json:"ssid"`
	Status       *netboxapi.Choice      `json:"status"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	AuthType     *netboxapi.Choice      `json:"auth_type"`
	AuthCipher   *netboxapi.Choice      `json:"auth_cipher"`
	AuthPsk      string                 `json:"auth_psk"`
	Distance     *float64               `json:"distance"`
	DistanceUnit *netboxapi.Choice      `json:"distance_unit"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func wirelessLinkRequestFromResourceData(api *providerState, d *schema.ResourceData) (*wirelessLinkRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &wirelessLinkRequest{
		InterfaceA:   int64ToPtr(int64(d.Get("interface_a_id").(int))),
		InterfaceB:   int64ToPtr(int64(d.Get("interface_b_id").(int))),
		Ssid:         getOptionalStr(d, "ssid", true),
		Status:       d.Get("status").(string),
		Tenant:       getOptionalInt(d, "tenant_id"),
		AuthType:     getOptionalStrClear(d, "auth_type"),
		AuthCipher:   getOptionalStrClear(d, "auth_cipher"),
		AuthPsk:      getOptionalStr(d, "auth_psk", true),
		Distance:     getOptionalFloat(d, "distance"),
		DistanceUnit: getOptionalStrClear(d, "distance_unit"),
		Description:  getOptionalStr(d, "description", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Comments:     getOptionalStr(d, "comments", true),
		Tags:         tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxWirelessLinkCreate(d *schema.ResourceData, m interface{}) error {
	data, err := wirelessLinkRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, wirelessLinkAPIPath, data, resourceNetboxWirelessLinkRead)
}

func resourceNetboxWirelessLinkRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, wirelessLinkAPIPath, func(r *wirelessLinkResponse) {
		d.Set("interface_a_id", netboxapi.RefID(r.InterfaceA))
		d.Set("interface_b_id", netboxapi.RefID(r.InterfaceB))
		d.Set("ssid", r.Ssid)
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("auth_type", netboxapi.ChoiceValue(r.AuthType))
		d.Set("auth_cipher", netboxapi.ChoiceValue(r.AuthCipher))
		d.Set("auth_psk", r.AuthPsk)
		d.Set("distance", r.Distance)
		d.Set("distance_unit", netboxapi.ChoiceValue(r.DistanceUnit))
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxWirelessLinkUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := wirelessLinkRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, wirelessLinkAPIPath, data, resourceNetboxWirelessLinkRead)
}

func resourceNetboxWirelessLinkDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, wirelessLinkAPIPath)
}
