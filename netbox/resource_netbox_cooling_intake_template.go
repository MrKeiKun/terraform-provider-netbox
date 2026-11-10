package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCoolingIntakeTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxCoolingIntakeTemplateCreate),
		ReadContext:   withContext(resourceNetboxCoolingIntakeTemplateRead),
		UpdateContext: withContext(resourceNetboxCoolingIntakeTemplateUpdate),
		DeleteContext: withContext(resourceNetboxCoolingIntakeTemplateDelete),

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/coolingintaketemplate/):

> A cooling intake template describes a liquid cooling inlet that is created on every device that is built from a device type or module type.

Requires NetBox v4.7.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"device_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device type associated with the cooling intake template.",
				Optional:    true,
			},
			"module_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module type associated with the cooling intake template.",
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "{module} is accepted as a substitution for the module bay position when attached to a module type.",
			},
			"label": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Physical label",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [uqd, uqdb, qdc, camlock, npt, bsp, proprietary]",
				ValidateFunc: validation.StringInSlice([]string{"uqd", "uqdb", "qdc", "camlock", "npt", "bsp", "proprietary"}, false),
			},
			"diameter": {
				Type:        schema.TypeFloat,
				Description: "The diameter of the cooling intake template.",
				Optional:    true,
			},
			"diameter_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [mm, cm, in]",
				ValidateFunc: validation.StringInSlice([]string{"mm", "cm", "in"}, false),
			},
			"max_flow": {
				Type:        schema.TypeFloat,
				Description: "The max flow of the cooling intake template.",
				Optional:    true,
			},
			"max_flow_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [lpm, m3ph, gpm]",
				ValidateFunc: validation.StringInSlice([]string{"lpm", "m3ph", "gpm"}, false),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cooling intake template.",
				Optional:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const coolingIntakeTemplateAPIPath = "dcim/cooling-intake-templates/"

type coolingIntakeTemplateRequest struct {
	DeviceType   *int64   `json:"device_type"`
	ModuleType   *int64   `json:"module_type"`
	Name         string   `json:"name"`
	Label        string   `json:"label,omitempty"`
	Type         *string  `json:"type,omitempty"`
	Diameter     *float64 `json:"diameter"`
	DiameterUnit *string  `json:"diameter_unit,omitempty"`
	MaxFlow      *float64 `json:"max_flow"`
	MaxFlowUnit  *string  `json:"max_flow_unit,omitempty"`
	Description  string   `json:"description,omitempty"`
}

type coolingIntakeTemplateResponse struct {
	DeviceType   *netboxapi.Ref    `json:"device_type"`
	ModuleType   *netboxapi.Ref    `json:"module_type"`
	Name         string            `json:"name"`
	Label        string            `json:"label"`
	Type         *netboxapi.Choice `json:"type"`
	Diameter     *float64          `json:"diameter"`
	DiameterUnit *netboxapi.Choice `json:"diameter_unit"`
	MaxFlow      *float64          `json:"max_flow"`
	MaxFlowUnit  *netboxapi.Choice `json:"max_flow_unit"`
	Description  string            `json:"description"`
}

func coolingIntakeTemplateRequestFromResourceData(api *providerState, d *schema.ResourceData) (*coolingIntakeTemplateRequest, error) {
	data := &coolingIntakeTemplateRequest{
		DeviceType:   getOptionalInt(d, "device_type_id"),
		ModuleType:   getOptionalInt(d, "module_type_id"),
		Name:         d.Get("name").(string),
		Label:        getOptionalStr(d, "label", true),
		Type:         getOptionalStrClear(d, "type"),
		Diameter:     getOptionalFloat(d, "diameter"),
		DiameterUnit: getOptionalStrClear(d, "diameter_unit"),
		MaxFlow:      getOptionalFloat(d, "max_flow"),
		MaxFlowUnit:  getOptionalStrClear(d, "max_flow_unit"),
		Description:  getOptionalStr(d, "description", true),
	}
	return data, nil
}

func resourceNetboxCoolingIntakeTemplateCreate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingIntakeTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, coolingIntakeTemplateAPIPath, data, resourceNetboxCoolingIntakeTemplateRead)
}

func resourceNetboxCoolingIntakeTemplateRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, coolingIntakeTemplateAPIPath, func(r *coolingIntakeTemplateResponse) {
		d.Set("device_type_id", netboxapi.RefID(r.DeviceType))
		d.Set("module_type_id", netboxapi.RefID(r.ModuleType))
		d.Set("name", r.Name)
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("diameter", r.Diameter)
		d.Set("diameter_unit", netboxapi.ChoiceValue(r.DiameterUnit))
		d.Set("max_flow", r.MaxFlow)
		d.Set("max_flow_unit", netboxapi.ChoiceValue(r.MaxFlowUnit))
		d.Set("description", r.Description)
	})
}

func resourceNetboxCoolingIntakeTemplateUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingIntakeTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, coolingIntakeTemplateAPIPath, data, resourceNetboxCoolingIntakeTemplateRead)
}

func resourceNetboxCoolingIntakeTemplateDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, coolingIntakeTemplateAPIPath)
}
