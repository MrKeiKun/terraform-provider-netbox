package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxModuleType() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxModuleTypeCreate,
		Read:   resourceNetboxModuleTypeRead,
		Update: resourceNetboxModuleTypeUpdate,
		Delete: resourceNetboxModuleTypeDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/moduletype/):

> A module type represents a specific make and model of hardware component which is installable within a device's module bay and has its own child components. For example, consider a chassis-based switch or router with a number of field-replaceable line cards. Each line card has its own model number and includes a certain set of components such as interfaces. Each module type may have a manufacturer, model number, and part number assigned to it.`,

		Schema: map[string]*schema.Schema{
			"manufacturer_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the manufacturer associated with the module type.",
				Required:    true,
			},
			"model": {
				Type:        schema.TypeString,
				Description: "The model of the module type.",
				Required:    true,
			},
			"part_number": {
				Type:        schema.TypeString,
				Description: "The part number of the module type.",
				Optional:    true,
			},
			"weight": {
				Type:        schema.TypeFloat,
				Description: "The weight of the module type.",
				Optional:    true,
			},
			"weight_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"weight"},
				Description:  "One of [kg, g, lb, oz]",
				ValidateFunc: validation.StringInSlice([]string{"kg", "g", "lb", "oz"}, false),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the module type.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the module type.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the module type.",
				Optional:    true,
			},
			"profile_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the profile associated with the module type.",
				Optional:    true,
			},
			"airflow": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [front-to-rear, rear-to-front, left-to-right, right-to-left, side-to-rear, passive]",
				ValidateFunc: validation.StringInSlice([]string{"front-to-rear", "rear-to-front", "left-to-right", "right-to-left", "side-to-rear", "passive"}, false),
			},
			"cooling_method": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [air, liquid, hybrid, immersion]",
				ValidateFunc: validation.StringInSlice([]string{"air", "liquid", "hybrid", "immersion"}, false),
			},
			"end_of_life": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The date after which this module type is no longer supported by the manufacturer",
			},
			"module_bay_type_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the module bay type associated with the module type.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const moduleTypeAPIPath = "dcim/module-types/"

type moduleTypeRequest struct {
	Manufacturer   *int64             `json:"manufacturer"`
	Model          string             `json:"model"`
	PartNumber     string             `json:"part_number,omitempty"`
	Weight         *float64           `json:"weight"`
	WeightUnit     string             `json:"weight_unit"`
	Description    string             `json:"description,omitempty"`
	Comments       string             `json:"comments,omitempty"`
	Owner          *int64             `json:"owner"`
	Profile        *int64             `json:"profile"`
	Airflow        *string            `json:"airflow,omitempty"`
	CoolingMethod  *string            `json:"cooling_method,omitempty"`
	EndOfLife      string             `json:"end_of_life,omitempty"`
	ModuleBayTypes []int64            `json:"module_bay_types"`
	CustomFields   interface{}        `json:"custom_fields,omitempty"`
	Tags           []netboxapi.TagRef `json:"tags"`
}

type moduleTypeResponse struct {
	Manufacturer   *netboxapi.Ref         `json:"manufacturer"`
	Model          string                 `json:"model"`
	PartNumber     string                 `json:"part_number"`
	Weight         *float64               `json:"weight"`
	WeightUnit     *netboxapi.Choice      `json:"weight_unit"`
	Description    string                 `json:"description"`
	Comments       string                 `json:"comments"`
	Owner          *netboxapi.Ref         `json:"owner"`
	Profile        *netboxapi.Ref         `json:"profile"`
	Airflow        *netboxapi.Choice      `json:"airflow"`
	CoolingMethod  *netboxapi.Choice      `json:"cooling_method"`
	EndOfLife      string                 `json:"end_of_life"`
	ModuleBayTypes []netboxapi.Ref        `json:"module_bay_types"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
	Tags           []netboxapi.TagRef     `json:"tags"`
}

func moduleTypeRequestFromResourceData(api *providerState, d *schema.ResourceData) (*moduleTypeRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &moduleTypeRequest{
		ModuleBayTypes: getIDSet(d, "module_bay_type_ids"),
		EndOfLife:      getOptionalStr(d, "end_of_life", true),
		CoolingMethod:  getOptionalStrClear(d, "cooling_method"),
		Airflow:        getOptionalStrClear(d, "airflow"),
		Profile:        getOptionalInt(d, "profile_id"),
		Owner:          getOptionalInt(d, "owner_id"),
		Manufacturer:   int64ToPtr(int64(d.Get("manufacturer_id").(int))),
		Model:          d.Get("model").(string),
		PartNumber:     getOptionalStr(d, "part_number", true),
		Weight:         getOptionalFloat(d, "weight"),
		WeightUnit:     getOptionalStr(d, "weight_unit", false),
		Description:    getOptionalStr(d, "description", true),
		Comments:       getOptionalStr(d, "comments", true),
		Tags:           tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxModuleTypeCreate(d *schema.ResourceData, m interface{}) error {
	data, err := moduleTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, moduleTypeAPIPath, data, resourceNetboxModuleTypeRead)
}

func resourceNetboxModuleTypeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, moduleTypeAPIPath, func(r *moduleTypeResponse) {
		d.Set("module_bay_type_ids", netboxapi.IDs(r.ModuleBayTypes))
		d.Set("end_of_life", r.EndOfLife)
		d.Set("cooling_method", netboxapi.ChoiceValue(r.CoolingMethod))
		d.Set("airflow", netboxapi.ChoiceValue(r.Airflow))
		d.Set("profile_id", netboxapi.RefID(r.Profile))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("manufacturer_id", netboxapi.RefID(r.Manufacturer))
		d.Set("model", r.Model)
		d.Set("part_number", r.PartNumber)
		d.Set("weight", r.Weight)
		d.Set("weight_unit", netboxapi.ChoiceValue(r.WeightUnit))
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxModuleTypeUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := moduleTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, moduleTypeAPIPath, data, resourceNetboxModuleTypeRead)
}

func resourceNetboxModuleTypeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, moduleTypeAPIPath)
}
