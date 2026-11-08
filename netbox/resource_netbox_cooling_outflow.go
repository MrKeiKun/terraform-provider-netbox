package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCoolingOutflow() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCoolingOutflowCreate,
		Read:   resourceNetboxCoolingOutflowRead,
		Update: resourceNetboxCoolingOutflowUpdate,
		Delete: resourceNetboxCoolingOutflowDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/coolingoutflow/):

> A cooling outflow is a liquid cooling outlet on a device.

Requires NetBox v4.7.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the cooling outflow.",
				Required:    true,
			},
			"module_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module associated with the cooling outflow.",
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the cooling outflow.",
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
				Description: "The diameter of the cooling outflow.",
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
				Description: "The ID of the cooling intake associated with the cooling outflow.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cooling outflow.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the cooling outflow.",
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

const coolingOutflowAPIPath = "dcim/cooling-outflows/"

type coolingOutflowRequest struct {
	Device        *int64             `json:"device"`
	Module        *int64             `json:"module"`
	Name          string             `json:"name"`
	Label         string             `json:"label,omitempty"`
	Type          *string            `json:"type,omitempty"`
	Diameter      *float64           `json:"diameter"`
	DiameterUnit  *string            `json:"diameter_unit,omitempty"`
	CoolingIntake *int64             `json:"cooling_intake"`
	Description   string             `json:"description,omitempty"`
	Owner         *int64             `json:"owner"`
	Tags          []netboxapi.TagRef `json:"tags"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
}

type coolingOutflowResponse struct {
	Device        *netboxapi.Ref         `json:"device"`
	Module        *netboxapi.Ref         `json:"module"`
	Name          string                 `json:"name"`
	Label         string                 `json:"label"`
	Type          *netboxapi.Choice      `json:"type"`
	Diameter      *float64               `json:"diameter"`
	DiameterUnit  *netboxapi.Choice      `json:"diameter_unit"`
	CoolingIntake *netboxapi.Ref         `json:"cooling_intake"`
	Description   string                 `json:"description"`
	Owner         *netboxapi.Ref         `json:"owner"`
	Tags          []netboxapi.TagRef     `json:"tags"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
}

func coolingOutflowRequestFromResourceData(api *providerState, d *schema.ResourceData) (*coolingOutflowRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &coolingOutflowRequest{
		Device:        int64ToPtr(int64(d.Get("device_id").(int))),
		Module:        getOptionalInt(d, "module_id"),
		Name:          d.Get("name").(string),
		Label:         getOptionalStr(d, "label", true),
		Type:          getOptionalStrClear(d, "type"),
		Diameter:      getOptionalFloat(d, "diameter"),
		DiameterUnit:  getOptionalStrClear(d, "diameter_unit"),
		CoolingIntake: getOptionalInt(d, "cooling_intake_id"),
		Description:   getOptionalStr(d, "description", true),
		Owner:         getOptionalInt(d, "owner_id"),
		Tags:          tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxCoolingOutflowCreate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingOutflowRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, coolingOutflowAPIPath, data, resourceNetboxCoolingOutflowRead)
}

func resourceNetboxCoolingOutflowRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, coolingOutflowAPIPath, func(r *coolingOutflowResponse) {
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("module_id", netboxapi.RefID(r.Module))
		d.Set("name", r.Name)
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("diameter", r.Diameter)
		d.Set("diameter_unit", netboxapi.ChoiceValue(r.DiameterUnit))
		d.Set("cooling_intake_id", netboxapi.RefID(r.CoolingIntake))
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxCoolingOutflowUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingOutflowRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, coolingOutflowAPIPath, data, resourceNetboxCoolingOutflowRead)
}

func resourceNetboxCoolingOutflowDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, coolingOutflowAPIPath)
}
