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
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const interfaceTemplateAPIPath = "dcim/interface-templates/"

type interfaceTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Label       string `json:"label,omitempty"`
	Type        string `json:"type"`
	MgmtOnly    bool   `json:"mgmt_only"`
	DeviceType  *int64 `json:"device_type"`
	ModuleType  *int64 `json:"module_type"`
}

type interfaceTemplateResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Label       string            `json:"label"`
	Type        *netboxapi.Choice `json:"type"`
	MgmtOnly    bool              `json:"mgmt_only"`
	DeviceType  *netboxapi.Ref    `json:"device_type"`
	ModuleType  *netboxapi.Ref    `json:"module_type"`
}

func interfaceTemplateRequestFromResourceData(d *schema.ResourceData) *interfaceTemplateRequest {
	name := d.Get("name").(string)
	data := &interfaceTemplateRequest{
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
