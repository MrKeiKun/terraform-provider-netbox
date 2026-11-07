package netbox

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxPowerPortTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxPowerPortTemplateCreate,
		ReadContext:   resourceNetboxPowerPortTemplateRead,
		UpdateContext: resourceNetboxPowerPortTemplateUpdate,
		DeleteContext: resourceNetboxPowerPortTemplateDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/powerporttemplate/):

> A template for a power port that will be created on all instantiations of the parent device type. See the power port documentation for more detail.`,
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
				Description: "One of [iec-60320-c6, iec-60320-c8, iec-60320-c14, iec-60320-c16, iec-60320-c20, iec-60320-c22, iec-60309-p-n-e-4h, iec-60309-p-n-e-6h, iec-60309-p-n-e-9h, iec-60309-2p-e-4h, iec-60309-2p-e-6h, iec-60309-2p-e-9h, iec-60309-3p-e-4h, iec-60309-3p-e-6h, iec-60309-3p-e-9h, iec-60309-3p-n-e-4h, iec-60309-3p-n-e-6h, iec-60309-3p-n-e-9h, nema-1-15p, nema-5-15p, nema-5-20p, nema-5-30p, nema-5-50p, nema-6-15p, nema-6-20p, nema-6-30p, nema-6-50p, nema-10-30p, nema-10-50p, nema-14-20p, nema-14-30p, nema-14-50p, nema-14-60p, nema-15-15p, nema-15-20p, nema-15-30p, nema-15-50p, nema-15-60p, nema-l1-15p, nema-l5-15p, nema-l5-20p, nema-l5-30p, nema-l5-50p, nema-l6-15p, nema-l6-20p, nema-l6-30p, nema-l6-50p, nema-l10-30p, nema-l14-20p, nema-l14-30p, nema-l14-50p, nema-l14-60p, nema-l15-20p, nema-l15-30p, nema-l15-50p, nema-l15-60p, nema-l21-20p, nema-l21-30p, nema-l22-30p, cs6361c, cs6365c, cs8165c, cs8265c, cs8365c, cs8465c, ita-c, ita-e, ita-f, ita-ef, ita-g, ita-h, ita-i, ita-j, ita-k, ita-l, ita-m, ita-n, ita-o, usb-a, usb-b, usb-c, usb-mini-a, usb-mini-b, usb-micro-a, usb-micro-b, usb-micro-ab, usb-3-b, usb-3-micro-b, dc-terminal, saf-d-grid, neutrik-powercon-20, neutrik-powercon-32, neutrik-powercon-true1, neutrik-powercon-true1-top, ubiquiti-smartpower, hardwired, other]",
			},
			"maximum_draw": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum power draw in watts.",
			},
			"allocated_draw": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Allocated power draw in watts.",
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

const powerPortTemplateAPIPath = "dcim/power-port-templates/"

type powerPortTemplateRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Label         string `json:"label,omitempty"`
	Type          string `json:"type,omitempty"`
	MaximumDraw   *int64 `json:"maximum_draw,omitempty"`
	AllocatedDraw *int64 `json:"allocated_draw,omitempty"`
	DeviceType    *int64 `json:"device_type"`
	ModuleType    *int64 `json:"module_type"`
}

type powerPortTemplateResponse struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Label         string            `json:"label"`
	Type          *netboxapi.Choice `json:"type"`
	MaximumDraw   *int64            `json:"maximum_draw"`
	AllocatedDraw *int64            `json:"allocated_draw"`
	DeviceType    *netboxapi.Ref    `json:"device_type"`
	ModuleType    *netboxapi.Ref    `json:"module_type"`
}

func powerPortTemplateRequestFromResourceData(d *schema.ResourceData) *powerPortTemplateRequest {
	name := d.Get("name").(string)
	data := &powerPortTemplateRequest{
		Name:          name,
		Description:   d.Get("description").(string),
		Label:         d.Get("label").(string),
		Type:          d.Get("type").(string),
		MaximumDraw:   getOptionalInt(d, "maximum_draw"),
		AllocatedDraw: getOptionalInt(d, "allocated_draw"),
	}
	if v := int64(d.Get("device_type_id").(int)); v != 0 || d.HasChange("device_type_id") {
		data.DeviceType = &v
	}
	if v := int64(d.Get("module_type_id").(int)); v != 0 || d.HasChange("module_type_id") {
		data.ModuleType = &v
	}
	return data
}

func resourceNetboxPowerPortTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createObjectCtx(ctx, d, m, powerPortTemplateAPIPath, powerPortTemplateRequestFromResourceData(d), resourceNetboxPowerPortTemplateRead)
}

func resourceNetboxPowerPortTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return readObjectCtx(ctx, d, m, powerPortTemplateAPIPath, func(r *powerPortTemplateResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("maximum_draw", r.MaximumDraw)
		d.Set("allocated_draw", r.AllocatedDraw)
		if r.DeviceType != nil {
			d.Set("device_type_id", r.DeviceType.ID)
		}
		if r.ModuleType != nil {
			d.Set("module_type_id", r.ModuleType.ID)
		}
	})
}

func resourceNetboxPowerPortTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return patchObjectCtx(ctx, d, m, powerPortTemplateAPIPath, powerPortTemplateRequestFromResourceData(d), resourceNetboxPowerPortTemplateRead)
}

func resourceNetboxPowerPortTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, powerPortTemplateAPIPath)
}
