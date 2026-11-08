package netbox

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxInterfaceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxInterfaceTemplateCreate,
		ReadContext:   resourceNetboxInterfaceTemplateRead,
		UpdateContext: resourceNetboxInterfaceTemplateUpdate,
		DeleteContext: resourceNetboxInterfaceTemplateDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/interfacetemplate/):

> A template for a network interface that will be created on all instantiations of the parent device type. See the interface documentation for more detail.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mgmt_only": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"device_type_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ExactlyOneOf: []string{"device_type_id", "module_type_id"},
				ForceNew:     true,
			},
			"module_type_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ExactlyOneOf: []string{"device_type_id", "module_type_id"},
				ForceNew:     true,
			},
			"channels": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 1024),
				Description:  "The number of channels into which this interface is channelized",
			},
			"channel_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 1024),
				Description:  "The channel on the parent interface to which this subinterface is bound",
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"bridge": {
				Type:     schema.TypeInt,
				Optional: true,
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
			"rf_role": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [ap, station]",
				ValidateFunc: validation.StringInSlice([]string{"ap", "station"}, false),
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const interfaceTemplateAPIPath = "dcim/interface-templates/"

type interfaceTemplateRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Label       string  `json:"label,omitempty"`
	Type        string  `json:"type"`
	MgmtOnly    bool    `json:"mgmt_only"`
	DeviceType  *int64  `json:"device_type"`
	ModuleType  *int64  `json:"module_type"`
	Channels    *int64  `json:"channels"`
	ChannelID   *int64  `json:"channel_id"`
	Enabled     *bool   `json:"enabled,omitempty"`
	Bridge      *int64  `json:"bridge"`
	PoeMode     *string `json:"poe_mode,omitempty"`
	PoeType     *string `json:"poe_type,omitempty"`
	RfRole      *string `json:"rf_role,omitempty"`
}

type interfaceTemplateResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Label       string            `json:"label"`
	Type        *netboxapi.Choice `json:"type"`
	MgmtOnly    bool              `json:"mgmt_only"`
	DeviceType  *netboxapi.Ref    `json:"device_type"`
	ModuleType  *netboxapi.Ref    `json:"module_type"`
	Channels    *int64            `json:"channels"`
	ChannelID   *int64            `json:"channel_id"`
	Enabled     bool              `json:"enabled"`
	Bridge      *int64            `json:"bridge"`
	PoeMode     *netboxapi.Choice `json:"poe_mode"`
	PoeType     *netboxapi.Choice `json:"poe_type"`
	RfRole      *netboxapi.Choice `json:"rf_role"`
}

func interfaceTemplateRequestFromResourceData(d *schema.ResourceData) *interfaceTemplateRequest {
	name := d.Get("name").(string)
	data := &interfaceTemplateRequest{
		RfRole:      getOptionalStrClear(d, "rf_role"),
		PoeType:     getOptionalStrClear(d, "poe_type"),
		PoeMode:     getOptionalStrClear(d, "poe_mode"),
		Bridge:      getOptionalInt(d, "bridge"),
		Enabled:     getOptionalBool(d, "enabled"),
		ChannelID:   getOptionalInt(d, "channel_id"),
		Channels:    getOptionalInt(d, "channels"),
		Name:        name,
		Description: d.Get("description").(string),
		Label:       d.Get("label").(string),
		Type:        d.Get("type").(string),
		MgmtOnly:    d.Get("mgmt_only").(bool),
	}
	if v := int64(d.Get("device_type_id").(int)); v != 0 || d.HasChange("device_type_id") {
		data.DeviceType = &v
	}
	if v := int64(d.Get("module_type_id").(int)); v != 0 || d.HasChange("module_type_id") {
		data.ModuleType = &v
	}
	return data
}

func resourceNetboxInterfaceTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createObjectCtx(ctx, d, m, interfaceTemplateAPIPath, interfaceTemplateRequestFromResourceData(d), resourceNetboxInterfaceTemplateRead)
}

func resourceNetboxInterfaceTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return readObjectCtx(ctx, d, m, interfaceTemplateAPIPath, func(r *interfaceTemplateResponse) {
		d.Set("rf_role", netboxapi.ChoiceValue(r.RfRole))
		d.Set("poe_type", netboxapi.ChoiceValue(r.PoeType))
		d.Set("poe_mode", netboxapi.ChoiceValue(r.PoeMode))
		d.Set("bridge", r.Bridge)
		d.Set("enabled", r.Enabled)
		d.Set("channel_id", r.ChannelID)
		d.Set("channels", r.Channels)
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("mgmt_only", r.MgmtOnly)
		if r.DeviceType != nil {
			d.Set("device_type_id", r.DeviceType.ID)
		}
		if r.ModuleType != nil {
			d.Set("module_type_id", r.ModuleType.ID)
		}
	})
}

func resourceNetboxInterfaceTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return patchObjectCtx(ctx, d, m, interfaceTemplateAPIPath, interfaceTemplateRequestFromResourceData(d), resourceNetboxInterfaceTemplateRead)
}

func resourceNetboxInterfaceTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, interfaceTemplateAPIPath)
}
