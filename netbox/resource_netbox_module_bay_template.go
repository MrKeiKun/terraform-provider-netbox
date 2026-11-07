package netbox

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxModuleBayTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxModuleBayTemplateCreate,
		ReadContext:   resourceNetboxModuleBayTemplateRead,
		UpdateContext: resourceNetboxModuleBayTemplateUpdate,
		DeleteContext: resourceNetboxModuleBayTemplateDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/modulebaytemplate/):

> A template for a module bay that will be created on all instantiations of the parent device type. See the module bay documentation for more detail.`,
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
			"position": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier to reference when renaming installed components.",
			},
			"device_type_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const moduleBayTemplateAPIPath = "dcim/module-bay-templates/"

type moduleBayTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Label       string `json:"label,omitempty"`
	Position    string `json:"position,omitempty"`
	DeviceType  *int64 `json:"device_type"`
}

type moduleBayTemplateResponse struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Label       string         `json:"label"`
	Position    string         `json:"position"`
	DeviceType  *netboxapi.Ref `json:"device_type"`
}

func moduleBayTemplateRequestFromResourceData(d *schema.ResourceData) *moduleBayTemplateRequest {
	name := d.Get("name").(string)
	data := &moduleBayTemplateRequest{
		Name:        name,
		Description: d.Get("description").(string),
		Label:       d.Get("label").(string),
		Position:    d.Get("position").(string),
		DeviceType:  int64ToPtr(int64(d.Get("device_type_id").(int))),
	}
	return data
}

func resourceNetboxModuleBayTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createObjectCtx(ctx, d, m, moduleBayTemplateAPIPath, moduleBayTemplateRequestFromResourceData(d), resourceNetboxModuleBayTemplateRead)
}

func resourceNetboxModuleBayTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return readObjectCtx(ctx, d, m, moduleBayTemplateAPIPath, func(r *moduleBayTemplateResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("label", r.Label)
		d.Set("position", r.Position)
		if r.DeviceType != nil {
			d.Set("device_type_id", r.DeviceType.ID)
		}
	})
}

func resourceNetboxModuleBayTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return patchObjectCtx(ctx, d, m, moduleBayTemplateAPIPath, moduleBayTemplateRequestFromResourceData(d), resourceNetboxModuleBayTemplateRead)
}

func resourceNetboxModuleBayTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, moduleBayTemplateAPIPath)
}
