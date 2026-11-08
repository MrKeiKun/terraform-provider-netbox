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
				Description:  "The name of the module bay template.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the module bay template.",
				Optional:    true,
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the module bay template.",
				Optional:    true,
			},
			"position": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier to reference when renaming installed components.",
			},
			"device_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device type associated with the module bay template.",
				Required:    true,
				ForceNew:    true,
			},
			"module_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module type associated with the module bay template.",
				Optional:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the module bay template is enabled.",
				Optional:    true,
				Computed:    true,
			},
			"module_bay_type_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the module bay type associated with the module bay template.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const moduleBayTemplateAPIPath = "dcim/module-bay-templates/"

type moduleBayTemplateRequest struct {
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	Label          string  `json:"label,omitempty"`
	Position       string  `json:"position,omitempty"`
	DeviceType     *int64  `json:"device_type"`
	ModuleType     *int64  `json:"module_type"`
	Enabled        *bool   `json:"enabled,omitempty"`
	ModuleBayTypes []int64 `json:"module_bay_types"`
}

type moduleBayTemplateResponse struct {
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Label          string          `json:"label"`
	Position       string          `json:"position"`
	DeviceType     *netboxapi.Ref  `json:"device_type"`
	ModuleType     *netboxapi.Ref  `json:"module_type"`
	Enabled        bool            `json:"enabled"`
	ModuleBayTypes []netboxapi.Ref `json:"module_bay_types"`
}

func moduleBayTemplateRequestFromResourceData(d *schema.ResourceData) *moduleBayTemplateRequest {
	name := d.Get("name").(string)
	data := &moduleBayTemplateRequest{
		ModuleBayTypes: getIDSet(d, "module_bay_type_ids"),
		Enabled:        getOptionalBool(d, "enabled"),
		ModuleType:     getOptionalInt(d, "module_type_id"),
		Name:           name,
		Description:    d.Get("description").(string),
		Label:          d.Get("label").(string),
		Position:       d.Get("position").(string),
		DeviceType:     int64ToPtr(int64(d.Get("device_type_id").(int))),
	}
	return data
}

func resourceNetboxModuleBayTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createObjectCtx(ctx, d, m, moduleBayTemplateAPIPath, moduleBayTemplateRequestFromResourceData(d), resourceNetboxModuleBayTemplateRead)
}

func resourceNetboxModuleBayTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return readObjectCtx(ctx, d, m, moduleBayTemplateAPIPath, func(r *moduleBayTemplateResponse) {
		d.Set("module_bay_type_ids", netboxapi.IDs(r.ModuleBayTypes))
		d.Set("enabled", r.Enabled)
		d.Set("module_type_id", netboxapi.RefID(r.ModuleType))
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
