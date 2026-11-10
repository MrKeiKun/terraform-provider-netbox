package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCoolingIntake() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxCoolingIntakeCreate),
		ReadContext:   withContext(resourceNetboxCoolingIntakeRead),
		UpdateContext: withContext(resourceNetboxCoolingIntakeUpdate),
		DeleteContext: withContext(resourceNetboxCoolingIntakeDelete),

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/coolingintake/):

> A cooling intake is a liquid cooling inlet on a device.

Requires NetBox v4.7.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the cooling intake.",
				Required:    true,
			},
			"module_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module associated with the cooling intake.",
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the cooling intake.",
				Required:    true,
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
				Description: "The diameter of the cooling intake.",
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
				Description: "The max flow of the cooling intake.",
				Optional:    true,
			},
			"max_flow_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [lpm, m3ph, gpm]",
				ValidateFunc: validation.StringInSlice([]string{"lpm", "m3ph", "gpm"}, false),
			},
			"cooling_outflow_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The upstream cooling outflow supplying this intake",
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cooling intake.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the cooling intake.",
				Optional:    true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const coolingIntakeAPIPath = "dcim/cooling-intakes/"

type coolingIntakeRequest struct {
	Device         *int64             `json:"device"`
	Module         *int64             `json:"module"`
	Name           string             `json:"name"`
	Label          string             `json:"label,omitempty"`
	Type           *string            `json:"type,omitempty"`
	Diameter       *float64           `json:"diameter"`
	DiameterUnit   *string            `json:"diameter_unit,omitempty"`
	MaxFlow        *float64           `json:"max_flow"`
	MaxFlowUnit    *string            `json:"max_flow_unit,omitempty"`
	CoolingOutflow *int64             `json:"cooling_outflow"`
	Description    string             `json:"description,omitempty"`
	Owner          *int64             `json:"owner"`
	Tags           []netboxapi.TagRef `json:"tags"`
	CustomFields   interface{}        `json:"custom_fields,omitempty"`
}

type coolingIntakeResponse struct {
	Device         *netboxapi.Ref         `json:"device"`
	Module         *netboxapi.Ref         `json:"module"`
	Name           string                 `json:"name"`
	Label          string                 `json:"label"`
	Type           *netboxapi.Choice      `json:"type"`
	Diameter       *float64               `json:"diameter"`
	DiameterUnit   *netboxapi.Choice      `json:"diameter_unit"`
	MaxFlow        *float64               `json:"max_flow"`
	MaxFlowUnit    *netboxapi.Choice      `json:"max_flow_unit"`
	CoolingOutflow *netboxapi.Ref         `json:"cooling_outflow"`
	Description    string                 `json:"description"`
	Owner          *netboxapi.Ref         `json:"owner"`
	Tags           []netboxapi.TagRef     `json:"tags"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
}

func coolingIntakeRequestFromResourceData(api *providerState, d *schema.ResourceData) (*coolingIntakeRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &coolingIntakeRequest{
		Device:         int64ToPtr(int64(d.Get("device_id").(int))),
		Module:         getOptionalInt(d, "module_id"),
		Name:           d.Get("name").(string),
		Label:          getOptionalStr(d, "label", true),
		Type:           getOptionalStrClear(d, "type"),
		Diameter:       getOptionalFloat(d, "diameter"),
		DiameterUnit:   getOptionalStrClear(d, "diameter_unit"),
		MaxFlow:        getOptionalFloat(d, "max_flow"),
		MaxFlowUnit:    getOptionalStrClear(d, "max_flow_unit"),
		CoolingOutflow: getOptionalInt(d, "cooling_outflow_id"),
		Description:    getOptionalStr(d, "description", true),
		Owner:          getOptionalInt(d, "owner_id"),
		Tags:           tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxCoolingIntakeCreate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingIntakeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, coolingIntakeAPIPath, data, resourceNetboxCoolingIntakeRead)
}

func resourceNetboxCoolingIntakeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, coolingIntakeAPIPath, func(r *coolingIntakeResponse) {
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("module_id", netboxapi.RefID(r.Module))
		d.Set("name", r.Name)
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("diameter", r.Diameter)
		d.Set("diameter_unit", netboxapi.ChoiceValue(r.DiameterUnit))
		d.Set("max_flow", r.MaxFlow)
		d.Set("max_flow_unit", netboxapi.ChoiceValue(r.MaxFlowUnit))
		d.Set("cooling_outflow_id", netboxapi.RefID(r.CoolingOutflow))
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxCoolingIntakeUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingIntakeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, coolingIntakeAPIPath, data, resourceNetboxCoolingIntakeRead)
}

func resourceNetboxCoolingIntakeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, coolingIntakeAPIPath)
}
