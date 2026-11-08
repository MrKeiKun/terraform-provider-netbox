package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxRackTypeFormFactorOptions = []string{"2-post-frame", "4-post-frame", "4-post-cabinet", "wall-frame", "wall-frame-vertical", "wall-cabinet", "wall-cabinet-vertical"}
var resourceNetboxRackTypeWeightUnitOptions = []string{"kg", "g", "lb", "oz"}
var resourceNetboxRackTypeOuterUnitOptions = []string{"mm", "in"}
var resourceNetboxRackTypeWidthOptions = []int{10, 19, 21, 23}

func resourceNetboxRackType() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxRackTypeCreate,
		Read:   resourceNetboxRackTypeRead,
		Update: resourceNetboxRackTypeUpdate,
		Delete: resourceNetboxRackTypeDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://netboxlabs.com/docs/netbox/en/stable/models/dcim/racktype/):

> A rack type defines the physical characteristics of a particular model of rack.`,

		Schema: map[string]*schema.Schema{
			"model": {
				Type:        schema.TypeString,
				Description: "The model of the rack type.",
				Required:    true,
			},
			"manufacturer_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the manufacturer associated with the rack type.",
				Optional:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the rack type.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"form_factor": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxRackTypeFormFactorOptions, false),
				Description:  buildValidValueDescription(resourceNetboxRackTypeFormFactorOptions),
			},
			"width": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntInSlice(resourceNetboxRackTypeWidthOptions),
				Description:  "Valid values are `10`, `19`, `21` and `23`",
			},
			"u_height": {
				Type:         schema.TypeInt,
				Description:  "The height of the rack type in rack units.",
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 100),
			},
			"starting_unit": {
				Type:        schema.TypeInt,
				Description: "The starting unit of the rack type.",
				Required:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the rack type.",
				Optional:    true,
			},
			"desc_units": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Units are numbered top-to-bottom",
			},
			"outer_height": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 32767),
				Description:  "Outer dimension of rack (height)",
			},
			"cooling_capability": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [air-only, hybrid, liquid-only]",
				ValidateFunc: validation.StringInSlice([]string{"air-only", "hybrid", "liquid-only"}, false),
			},
			"cooling_capacity": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Cooling capacity (kW)",
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
			"description": {
				Type:         schema.TypeString,
				Description:  "A description of the rack type.",
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"outer_width": {
				Type:         schema.TypeInt,
				Description:  "The outer width of the rack type.",
				Optional:     true,
				ValidateFunc: validatePositiveInt16,
			},
			"outer_depth": {
				Type:         schema.TypeInt,
				Description:  "The outer depth of the rack type.",
				Optional:     true,
				ValidateFunc: validatePositiveInt16,
			},
			"outer_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"outer_width", "outer_depth"},
				ValidateFunc: validation.StringInSlice(resourceNetboxRackTypeOuterUnitOptions, false),
				Description:  buildValidValueDescription(resourceNetboxRackTypeOuterUnitOptions),
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the rack type.",
				Optional:    true,
			},
			"weight": {
				Type:        schema.TypeFloat,
				Description: "The weight of the rack type.",
				Optional:    true,
			},
			"max_weight": {
				Type:         schema.TypeInt,
				Description:  "The max weight of the rack type.",
				Optional:     true,
				ValidateFunc: validatePositiveInt32,
			},
			"weight_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"weight", "max_weight"},
				ValidateFunc: validation.StringInSlice(resourceNetboxRackTypeWeightUnitOptions, false),
				Description:  buildValidValueDescription(resourceNetboxRackTypeWeightUnitOptions),
			},
			"mounting_depth_mm": {
				Type:         schema.TypeInt,
				Description:  "The mounting depth mm of the rack type.",
				Optional:     true,
				ValidateFunc: validatePositiveInt16,
			},
			//			"tenant_id": {
			//				Type:     schema.TypeInt,
			//				Optional: true,
			//			},
			//			"facility_id": {
			//				Type:     schema.TypeString,
			//				Optional: true,
			//			},
			//			"location_id": {
			//				Type:     schema.TypeInt,
			//				Optional: true,
			//			},
			//			"role_id": {
			//				Type:     schema.TypeInt,
			//				Optional: true,
			//			},
			//			"serial": {
			//				Type:         schema.TypeString,
			//				Optional:     true,
			//				ValidateFunc: validation.StringLenBetween(0, 50),
			//			},
			//			"asset_tag": {
			//				Type:         schema.TypeString,
			//				Optional:     true,
			//				ValidateFunc: validation.StringLenBetween(0, 50),
			//			},
			//			"type": {
			//				Type:         schema.TypeString,
			//				Optional:     true,
			//				ValidateFunc: validation.StringInSlice(resourceNetboxRackTypeTypeOptions, false),
			//				Description:  buildValidValueDescription(resourceNetboxRackTypeTypeOptions),
			//			},
			//			"desc_units": {
			//				Type:        schema.TypeBool,
			//				Optional:    true,
			//				Description: "If rack units are descending",
			//				Default:     false,
			//			},
			//			"outer_width": {
			//				Type:         schema.TypeInt,
			//				Optional:     true,
			//				ValidateFunc: validatePositiveInt16,
			//			},
			//			"outer_depth": {
			//				Type:         schema.TypeInt,
			//				Optional:     true,
			//				ValidateFunc: validatePositiveInt16,
			//			},
			//			"outer_unit": {
			//				Type:         schema.TypeString,
			//				Optional:     true,
			//				RequiredWith: []string{"outer_width", "outer_depth"},
			//				ValidateFunc: validation.StringInSlice(resourceNetboxRackTypeOuterUnitOptions, false),
			//				Description:  buildValidValueDescription(resourceNetboxRackTypeOuterUnitOptions),
			//			},
			//			"mounting_depth": {
			//				Type:         schema.TypeInt,
			//				Optional:     true,
			//				ValidateFunc: validatePositiveInt16,
			//			},
			//			"description": {
			//				Type:         schema.TypeString,
			//				Optional:     true,
			//				ValidateFunc: validation.StringLenBetween(0, 200),
			//			},
			//			"comments": {
			//				Type:     schema.TypeString,
			//				Optional: true,
			//			},
			//			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const rackTypeAPIPath = "dcim/rack-types/"

type rackTypeRequest struct {
	Slug              string             `json:"slug"`
	Model             string             `json:"model"`
	FormFactor        string             `json:"form_factor"`
	Manufacturer      *int64             `json:"manufacturer"`
	Width             *int64             `json:"width,omitempty"`
	UHeight           int64              `json:"u_height,omitempty"`
	Description       string             `json:"description,omitempty"`
	OuterWidth        *int64             `json:"outer_width,omitempty"`
	OuterDepth        *int64             `json:"outer_depth,omitempty"`
	OuterUnit         string             `json:"outer_unit,omitempty"`
	Comments          string             `json:"comments,omitempty"`
	Weight            *float64           `json:"weight,omitempty"`
	MaxWeight         *int64             `json:"max_weight,omitempty"`
	WeightUnit        string             `json:"weight_unit,omitempty"`
	MountingDepth     *int64             `json:"mounting_depth,omitempty"`
	Owner             *int64             `json:"owner"`
	DescUnits         *bool              `json:"desc_units,omitempty"`
	OuterHeight       *int64             `json:"outer_height"`
	CoolingCapability *string            `json:"cooling_capability,omitempty"`
	CoolingCapacity   *float64           `json:"cooling_capacity"`
	CustomFields      interface{}        `json:"custom_fields,omitempty"`
	Tags              []netboxapi.TagRef `json:"tags"`
}

type rackTypeResponse struct {
	Slug              string                 `json:"slug"`
	Model             string                 `json:"model"`
	FormFactor        *netboxapi.Choice      `json:"form_factor"`
	Manufacturer      *netboxapi.Ref         `json:"manufacturer"`
	Width             *netboxapi.IntChoice   `json:"width"`
	UHeight           int64                  `json:"u_height"`
	Description       string                 `json:"description"`
	OuterWidth        *int64                 `json:"outer_width"`
	OuterDepth        *int64                 `json:"outer_depth"`
	OuterUnit         *netboxapi.Choice      `json:"outer_unit"`
	Comments          string                 `json:"comments"`
	Weight            *float64               `json:"weight"`
	MaxWeight         *int64                 `json:"max_weight"`
	WeightUnit        *netboxapi.Choice      `json:"weight_unit"`
	MountingDepth     *int64                 `json:"mounting_depth"`
	StartingUnit      int64                  `json:"starting_unit"`
	Owner             *netboxapi.Ref         `json:"owner"`
	DescUnits         bool                   `json:"desc_units"`
	OuterHeight       *int64                 `json:"outer_height"`
	CoolingCapability *netboxapi.Choice      `json:"cooling_capability"`
	CoolingCapacity   *float64               `json:"cooling_capacity"`
	CustomFields      map[string]interface{} `json:"custom_fields"`
	Tags              []netboxapi.TagRef     `json:"tags"`
}

func rackTypeRequestFromResourceData(api *providerState, d *schema.ResourceData) (*rackTypeRequest, error) {
	slugSource := d.Get("model").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &rackTypeRequest{
		CustomFields:      customFieldsRequest(d),
		CoolingCapacity:   getOptionalFloat(d, "cooling_capacity"),
		CoolingCapability: getOptionalStrClear(d, "cooling_capability"),
		OuterHeight:       getOptionalInt(d, "outer_height"),
		DescUnits:         getOptionalBool(d, "desc_units"),
		Owner:             getOptionalInt(d, "owner_id"),
		Slug:              slugFromResourceData(d, slugSource),
		Model:             d.Get("model").(string),
		FormFactor:        d.Get("form_factor").(string),
		Manufacturer:      int64ToPtr(int64(d.Get("manufacturer_id").(int))),
		Width:             getOptionalInt(d, "width"),
		UHeight:           int64(d.Get("u_height").(int)),
		Description:       getOptionalStr(d, "description", true),
		OuterWidth:        getOptionalInt(d, "outer_width"),
		OuterDepth:        getOptionalInt(d, "outer_depth"),
		OuterUnit:         d.Get("outer_unit").(string),
		Comments:          getOptionalStr(d, "comments", true),
		Weight:            getOptionalFloat(d, "weight"),
		MaxWeight:         getOptionalInt(d, "max_weight"),
		WeightUnit:        d.Get("weight_unit").(string),
		MountingDepth:     getOptionalInt(d, "mounting_depth_mm"),
		Tags:              tags,
	}
	return data, nil
}

func resourceNetboxRackTypeCreate(d *schema.ResourceData, m interface{}) error {
	data, err := rackTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, rackTypeAPIPath, data, resourceNetboxRackTypeRead)
}

func resourceNetboxRackTypeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, rackTypeAPIPath, func(r *rackTypeResponse) {
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("cooling_capacity", r.CoolingCapacity)
		d.Set("cooling_capability", netboxapi.ChoiceValue(r.CoolingCapability))
		d.Set("outer_height", r.OuterHeight)
		d.Set("desc_units", r.DescUnits)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("slug", r.Slug)
		d.Set("model", r.Model)
		d.Set("form_factor", netboxapi.ChoiceValue(r.FormFactor))
		d.Set("manufacturer_id", netboxapi.RefID(r.Manufacturer))
		if r.Width != nil {
			d.Set("width", r.Width.Value)
		} else {
			d.Set("width", nil)
		}
		d.Set("u_height", r.UHeight)
		d.Set("description", r.Description)
		d.Set("outer_width", r.OuterWidth)
		d.Set("outer_depth", r.OuterDepth)
		d.Set("outer_unit", netboxapi.ChoiceValue(r.OuterUnit))
		d.Set("comments", r.Comments)
		d.Set("weight", r.Weight)
		d.Set("max_weight", r.MaxWeight)
		d.Set("weight_unit", netboxapi.ChoiceValue(r.WeightUnit))
		d.Set("mounting_depth_mm", r.MountingDepth)

		d.Set("starting_unit", r.StartingUnit)
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxRackTypeUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := rackTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, rackTypeAPIPath, data, resourceNetboxRackTypeRead)
}

func resourceNetboxRackTypeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, rackTypeAPIPath)
}
