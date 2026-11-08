package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCoolingOutflowTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCoolingOutflowTemplateCreate,
		Read:   resourceNetboxCoolingOutflowTemplateRead,
		Update: resourceNetboxCoolingOutflowTemplateUpdate,
		Delete: resourceNetboxCoolingOutflowTemplateDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/coolingoutflowtemplate/):

> A cooling outflow template describes a liquid cooling outlet that is created on every device that is built from a device type or module type.

Requires NetBox v4.7.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"device_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device type associated with the cooling outflow template.",
				Optional:    true,
			},
			"module_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module type associated with the cooling outflow template.",
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
				Description: "The diameter of the cooling outflow template.",
				Optional:    true,
			},
			"diameter_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [mm, cm, in]",
				ValidateFunc: validation.StringInSlice([]string{"mm", "cm", "in"}, false),
			},
			"cooling_intake_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the cooling intake associated with the cooling outflow template.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cooling outflow template.",
				Optional:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const coolingOutflowTemplateAPIPath = "dcim/cooling-outflow-templates/"

type coolingOutflowTemplateRequest struct {
	DeviceType    *int64   `json:"device_type"`
	ModuleType    *int64   `json:"module_type"`
	Name          string   `json:"name"`
	Label         string   `json:"label,omitempty"`
	Type          *string  `json:"type,omitempty"`
	Diameter      *float64 `json:"diameter"`
	DiameterUnit  *string  `json:"diameter_unit,omitempty"`
	CoolingIntake *int64   `json:"cooling_intake"`
	Description   string   `json:"description,omitempty"`
}

type coolingOutflowTemplateResponse struct {
	DeviceType    *netboxapi.Ref    `json:"device_type"`
	ModuleType    *netboxapi.Ref    `json:"module_type"`
	Name          string            `json:"name"`
	Label         string            `json:"label"`
	Type          *netboxapi.Choice `json:"type"`
	Diameter      *float64          `json:"diameter"`
	DiameterUnit  *netboxapi.Choice `json:"diameter_unit"`
	CoolingIntake *netboxapi.Ref    `json:"cooling_intake"`
	Description   string            `json:"description"`
}

func coolingOutflowTemplateRequestFromResourceData(api *providerState, d *schema.ResourceData) (*coolingOutflowTemplateRequest, error) {
	data := &coolingOutflowTemplateRequest{
		DeviceType:    getOptionalInt(d, "device_type_id"),
		ModuleType:    getOptionalInt(d, "module_type_id"),
		Name:          d.Get("name").(string),
		Label:         getOptionalStr(d, "label", true),
		Type:          getOptionalStrClear(d, "type"),
		Diameter:      getOptionalFloat(d, "diameter"),
		DiameterUnit:  getOptionalStrClear(d, "diameter_unit"),
		CoolingIntake: getOptionalInt(d, "cooling_intake_id"),
		Description:   getOptionalStr(d, "description", true),
	}
	return data, nil
}

func resourceNetboxCoolingOutflowTemplateCreate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingOutflowTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, coolingOutflowTemplateAPIPath, data, resourceNetboxCoolingOutflowTemplateRead)
}

func resourceNetboxCoolingOutflowTemplateRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, coolingOutflowTemplateAPIPath, func(r *coolingOutflowTemplateResponse) {
		d.Set("device_type_id", netboxapi.RefID(r.DeviceType))
		d.Set("module_type_id", netboxapi.RefID(r.ModuleType))
		d.Set("name", r.Name)
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("diameter", r.Diameter)
		d.Set("diameter_unit", netboxapi.ChoiceValue(r.DiameterUnit))
		d.Set("cooling_intake_id", netboxapi.RefID(r.CoolingIntake))
		d.Set("description", r.Description)
	})
}

func resourceNetboxCoolingOutflowTemplateUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingOutflowTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, coolingOutflowTemplateAPIPath, data, resourceNetboxCoolingOutflowTemplateRead)
}

func resourceNetboxCoolingOutflowTemplateDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, coolingOutflowTemplateAPIPath)
}
