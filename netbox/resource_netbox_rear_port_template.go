package netbox

import (
	"context"
	"regexp"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxRearPortTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxRearPortTemplateCreate,
		ReadContext:   resourceNetboxRearPortTemplateRead,
		UpdateContext: resourceNetboxRearPortTemplateUpdate,
		DeleteContext: resourceNetboxRearPortTemplateDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/rearporttemplate/):

> A template for a rear port that will be created on all instantiations of the parent device type. See the rear port documentation for more detail.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the rear port template.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the rear port template.",
				Optional:    true,
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the rear port template.",
				Optional:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "One of [8p8c, 8p6c, 8p4c, 8p2c, 6p6c, 6p4c, 6p2c, 4p4c, 4p2c, gg45, tera-4p, tera-2p, tera-1p, 110-punch, bnc, f, n, mrj21, fc, lc, lc-pc, lc-upc, lc-apc, lsh, lsh-pc, lsh-upc, lsh-apc, mpo, mtrj, sc, sc-pc, sc-upc, sc-apc, st, cs, sn, sma-905, sma-906, urm-p2, urm-p4, urm-p8, splice, other]",
			},
			"positions": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 1024),
				Description:  "Number of front ports which may be mapped to this rear port.",
			},
			"color_hex": {
				Type:         schema.TypeString,
				Description:  "The color hex of the rear port template.",
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[0-9a-f]{6}$"), "Must be hex color string"),
			},
			"device_type_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the device type associated with the rear port template.",
				Optional:     true,
				ExactlyOneOf: []string{"device_type_id", "module_type_id"},
				ForceNew:     true,
			},
			"module_type_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the module type associated with the rear port template.",
				Optional:     true,
				ExactlyOneOf: []string{"device_type_id", "module_type_id"},
				ForceNew:     true,
			},
			"front_ports": portMappingsSchema("front_port"),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const rearPortTemplateAPIPath = "dcim/rear-port-templates/"

type rearPortTemplateRequest struct {
	FrontPorts  []map[string]int64 `json:"front_ports,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Label       string             `json:"label,omitempty"`
	Type        string             `json:"type"`
	Positions   int64              `json:"positions,omitempty"`
	Color       string             `json:"color,omitempty"`
	DeviceType  *int64             `json:"device_type"`
	ModuleType  *int64             `json:"module_type"`
}

type rearPortTemplateResponse struct {
	FrontPorts  []map[string]interface{} `json:"front_ports"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Label       string                   `json:"label"`
	Type        *netboxapi.Choice        `json:"type"`
	Positions   int64                    `json:"positions"`
	Color       string                   `json:"color"`
	DeviceType  *netboxapi.Ref           `json:"device_type"`
	ModuleType  *netboxapi.Ref           `json:"module_type"`
}

func rearPortTemplateRequestFromResourceData(d *schema.ResourceData) *rearPortTemplateRequest {
	name := d.Get("name").(string)
	data := &rearPortTemplateRequest{
		FrontPorts:  portMappingsRequest(d, "front_ports", "front_port"),
		Name:        name,
		Description: d.Get("description").(string),
		Label:       d.Get("label").(string),
		Type:        d.Get("type").(string),
		Positions:   int64(d.Get("positions").(int)),
		Color:       d.Get("color_hex").(string),
	}
	if v := int64(d.Get("device_type_id").(int)); v != 0 || d.HasChange("device_type_id") {
		data.DeviceType = &v
	}
	if v := int64(d.Get("module_type_id").(int)); v != 0 || d.HasChange("module_type_id") {
		data.ModuleType = &v
	}
	// NetBox rejects an update that sends the mapping it already has, so the
	// mapping is only sent when it changed.
	if d.Id() != "" && !d.HasChange("front_ports") {
		data.FrontPorts = nil
	}
	return data
}

func resourceNetboxRearPortTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createObjectCtx(ctx, d, m, rearPortTemplateAPIPath, rearPortTemplateRequestFromResourceData(d), resourceNetboxRearPortTemplateRead)
}

func resourceNetboxRearPortTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return readObjectCtx(ctx, d, m, rearPortTemplateAPIPath, func(r *rearPortTemplateResponse) {
		d.Set("front_ports", portMappingsState(r.FrontPorts, "front_port"))
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("positions", r.Positions)
		d.Set("color_hex", r.Color)
		if r.DeviceType != nil {
			d.Set("device_type_id", r.DeviceType.ID)
		}
		if r.ModuleType != nil {
			d.Set("module_type_id", r.ModuleType.ID)
		}
	})
}

func resourceNetboxRearPortTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return patchObjectCtx(ctx, d, m, rearPortTemplateAPIPath, rearPortTemplateRequestFromResourceData(d), resourceNetboxRearPortTemplateRead)
}

func resourceNetboxRearPortTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, rearPortTemplateAPIPath)
}
