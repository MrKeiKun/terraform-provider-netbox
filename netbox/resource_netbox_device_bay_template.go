package netbox

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxDeviceBayTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxDeviceBayTemplateCreate,
		ReadContext:   resourceNetboxDeviceBayTemplateRead,
		UpdateContext: resourceNetboxDeviceBayTemplateUpdate,
		DeleteContext: resourceNetboxDeviceBayTemplateDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/devicebaytemplate/):

> A template for a device bay that will be created on all instantiations of the parent device type.`,
		Schema: map[string]*schema.Schema{
			"device_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device type associated with the device bay template.",
				Required:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the device bay template.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the device bay template.",
				Optional:    true,
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the device bay template.",
				Optional:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the device bay template is enabled.",
				Optional:    true,
				Computed:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const deviceBayTemplateAPIPath = "dcim/device-bay-templates/"

type deviceBayTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Label       string `json:"label,omitempty"`
	DeviceType  *int64 `json:"device_type"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

type deviceBayTemplateResponse struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Label       string         `json:"label"`
	DeviceType  *netboxapi.Ref `json:"device_type"`
	Enabled     bool           `json:"enabled"`
}

func deviceBayTemplateRequestFromResourceData(d *schema.ResourceData) *deviceBayTemplateRequest {
	name := d.Get("name").(string)
	data := &deviceBayTemplateRequest{
		Enabled:     getOptionalBool(d, "enabled"),
		Name:        name,
		Description: d.Get("description").(string),
		Label:       d.Get("label").(string),
	}
	if v := int64(d.Get("device_type_id").(int)); v != 0 || d.HasChange("device_type_id") {
		data.DeviceType = &v
	}
	return data
}

func resourceNetboxDeviceBayTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createObjectCtx(ctx, d, m, deviceBayTemplateAPIPath, deviceBayTemplateRequestFromResourceData(d), resourceNetboxDeviceBayTemplateRead)
}

func resourceNetboxDeviceBayTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return readObjectCtx(ctx, d, m, deviceBayTemplateAPIPath, func(r *deviceBayTemplateResponse) {
		d.Set("enabled", r.Enabled)
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("label", r.Label)
		if r.DeviceType != nil {
			d.Set("device_type_id", r.DeviceType.ID)
		}
	})
}

func resourceNetboxDeviceBayTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return patchObjectCtx(ctx, d, m, deviceBayTemplateAPIPath, deviceBayTemplateRequestFromResourceData(d), resourceNetboxDeviceBayTemplateRead)
}

func resourceNetboxDeviceBayTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, deviceBayTemplateAPIPath)
}
