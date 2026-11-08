package netbox

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxConsoleServerPortTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxConsoleServerPortTemplateCreate,
		ReadContext:   resourceNetboxConsoleServerPortTemplateRead,
		UpdateContext: resourceNetboxConsoleServerPortTemplateUpdate,
		DeleteContext: resourceNetboxConsoleServerPortTemplateDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/consoleserverporttemplate/):

> A template for a console server port that will be created on all instantiations of the parent device type. See the console server port documentation for more detail.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the console server port template.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the console server port template.",
				Optional:    true,
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the console server port template.",
				Optional:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "One of [de-9, db-25, rj-11, rj-12, rj-45, mini-din-8, usb-a, usb-b, usb-c, usb-mini-a, usb-mini-b, usb-micro-a, usb-micro-b, usb-micro-ab, other]",
			},
			"device_type_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the device type associated with the console server port template.",
				Optional:     true,
				ExactlyOneOf: []string{"device_type_id", "module_type_id"},
				ForceNew:     true,
			},
			"module_type_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the module type associated with the console server port template.",
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

const consoleServerPortTemplateAPIPath = "dcim/console-server-port-templates/"

type consoleServerPortTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Label       string `json:"label,omitempty"`
	Type        string `json:"type,omitempty"`
	DeviceType  *int64 `json:"device_type"`
	ModuleType  *int64 `json:"module_type"`
}

type consoleServerPortTemplateResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Label       string            `json:"label"`
	Type        *netboxapi.Choice `json:"type"`
	DeviceType  *netboxapi.Ref    `json:"device_type"`
	ModuleType  *netboxapi.Ref    `json:"module_type"`
}

func consoleServerPortTemplateRequestFromResourceData(d *schema.ResourceData) *consoleServerPortTemplateRequest {
	name := d.Get("name").(string)
	data := &consoleServerPortTemplateRequest{
		Name:        name,
		Description: d.Get("description").(string),
		Label:       d.Get("label").(string),
		Type:        d.Get("type").(string),
	}
	if v := int64(d.Get("device_type_id").(int)); v != 0 || d.HasChange("device_type_id") {
		data.DeviceType = &v
	}
	if v := int64(d.Get("module_type_id").(int)); v != 0 || d.HasChange("module_type_id") {
		data.ModuleType = &v
	}
	return data
}

func resourceNetboxConsoleServerPortTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createObjectCtx(ctx, d, m, consoleServerPortTemplateAPIPath, consoleServerPortTemplateRequestFromResourceData(d), resourceNetboxConsoleServerPortTemplateRead)
}

func resourceNetboxConsoleServerPortTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return readObjectCtx(ctx, d, m, consoleServerPortTemplateAPIPath, func(r *consoleServerPortTemplateResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		if r.DeviceType != nil {
			d.Set("device_type_id", r.DeviceType.ID)
		}
		if r.ModuleType != nil {
			d.Set("module_type_id", r.ModuleType.ID)
		}
	})
}

func resourceNetboxConsoleServerPortTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return patchObjectCtx(ctx, d, m, consoleServerPortTemplateAPIPath, consoleServerPortTemplateRequestFromResourceData(d), resourceNetboxConsoleServerPortTemplateRead)
}

func resourceNetboxConsoleServerPortTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, consoleServerPortTemplateAPIPath)
}
