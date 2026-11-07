package netbox

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxPowerOutletTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxPowerOutletTemplateCreate,
		ReadContext:   resourceNetboxPowerOutletTemplateRead,
		UpdateContext: resourceNetboxPowerOutletTemplateUpdate,
		DeleteContext: resourceNetboxPowerOutletTemplateDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/poweroutlettemplate/):

> A template for a power outlet that will be created on all instantiations of the parent device type. See the power outlet documentation for more detail.`,
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
				Type:        schema.TypeString,
				Optional:    true,
				Description: "One of [iec-60320-c5, iec-60320-c7, iec-60320-c13, iec-60320-c15, iec-60320-c19, iec-60320-c21, iec-60309-p-n-e-4h, iec-60309-p-n-e-6h, iec-60309-p-n-e-9h, iec-60309-2p-e-4h, iec-60309-2p-e-6h, iec-60309-2p-e-9h, iec-60309-3p-e-4h, iec-60309-3p-e-6h, iec-60309-3p-e-9h, iec-60309-3p-n-e-4h, iec-60309-3p-n-e-6h, iec-60309-3p-n-e-9h, nema-1-15r, nema-5-15r, nema-5-20r, nema-5-30r, nema-5-50r, nema-6-15r, nema-6-20r, nema-6-30r, nema-6-50r, nema-10-30r, nema-10-50r, nema-14-20r, nema-14-30r, nema-14-50r, nema-14-60r, nema-15-15r, nema-15-20r, nema-15-30r, nema-15-50r, nema-15-60r, nema-l1-15r, nema-l5-15r, nema-l5-20r, nema-l5-30r, nema-l5-50r, nema-l6-15r, nema-l6-20r, nema-l6-30r, nema-l6-50r, nema-l10-30r, nema-l14-20r, nema-l14-30r, nema-l14-50r, nema-l14-60r, nema-l15-20r, nema-l15-30r, nema-l15-50r, nema-l15-60r, nema-l21-20r, nema-l21-30r, nema-l22-30r, CS6360C, CS6364C, CS8164C, CS8264C, CS8364C, CS8464C, ita-e, ita-f, ita-g, ita-h, ita-i, ita-j, ita-k, ita-l, ita-m, ita-n, ita-o, ita-multistandard, usb-a, usb-micro-b, usb-c, dc-terminal, hdot-cx, saf-d-grid, neutrik-powercon-20a, neutrik-powercon-32a, neutrik-powercon-true1, neutrik-powercon-true1-top, ubiquiti-smartpower, hardwired, other]",
			},
			"power_port_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ID of a power port template on the same parent device type or module type.",
			},
			"feed_leg": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [A, B, C]",
				ValidateFunc: validation.StringInSlice([]string{"A", "B", "C"}, false),
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
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const powerOutletTemplateAPIPath = "dcim/power-outlet-templates/"

type powerOutletTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Label       string `json:"label,omitempty"`
	Type        string `json:"type,omitempty"`
	PowerPort   *int64 `json:"power_port,omitempty"`
	FeedLeg     string `json:"feed_leg,omitempty"`
	DeviceType  *int64 `json:"device_type"`
	ModuleType  *int64 `json:"module_type"`
}

type powerOutletTemplateResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Label       string            `json:"label"`
	Type        *netboxapi.Choice `json:"type"`
	PowerPort   *netboxapi.Ref    `json:"power_port"`
	FeedLeg     *netboxapi.Choice `json:"feed_leg"`
	DeviceType  *netboxapi.Ref    `json:"device_type"`
	ModuleType  *netboxapi.Ref    `json:"module_type"`
}

func powerOutletTemplateRequestFromResourceData(d *schema.ResourceData) *powerOutletTemplateRequest {
	name := d.Get("name").(string)
	data := &powerOutletTemplateRequest{
		Name:        name,
		Description: d.Get("description").(string),
		Label:       d.Get("label").(string),
		Type:        d.Get("type").(string),
		PowerPort:   getOptionalInt(d, "power_port_id"),
		FeedLeg:     d.Get("feed_leg").(string),
	}
	if v := int64(d.Get("device_type_id").(int)); v != 0 || d.HasChange("device_type_id") {
		data.DeviceType = &v
	}
	if v := int64(d.Get("module_type_id").(int)); v != 0 || d.HasChange("module_type_id") {
		data.ModuleType = &v
	}
	return data
}

func resourceNetboxPowerOutletTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createObjectCtx(ctx, d, m, powerOutletTemplateAPIPath, powerOutletTemplateRequestFromResourceData(d), resourceNetboxPowerOutletTemplateRead)
}

func resourceNetboxPowerOutletTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return readObjectCtx(ctx, d, m, powerOutletTemplateAPIPath, func(r *powerOutletTemplateResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		if r.PowerPort != nil {
			d.Set("power_port_id", r.PowerPort.ID)
		}
		d.Set("feed_leg", netboxapi.ChoiceValue(r.FeedLeg))
		if r.DeviceType != nil {
			d.Set("device_type_id", r.DeviceType.ID)
		}
		if r.ModuleType != nil {
			d.Set("module_type_id", r.ModuleType.ID)
		}
	})
}

func resourceNetboxPowerOutletTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return patchObjectCtx(ctx, d, m, powerOutletTemplateAPIPath, powerOutletTemplateRequestFromResourceData(d), resourceNetboxPowerOutletTemplateRead)
}

func resourceNetboxPowerOutletTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, powerOutletTemplateAPIPath)
}
