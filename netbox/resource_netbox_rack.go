package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxRackStatusOptions = []string{"reserved", "available", "planned", "active", "deprecated"}
var resourceNetboxRackWeightUnitOptions = []string{"kg", "g", "lb", "oz"}
var resourceNetboxRackOuterUnitOptions = []string{"mm", "in"}
var resourceNetboxRackWidthOptions = []int{10, 19, 21, 23}
var resourceNetboxRackFormFactorOptions = []string{"2-post-frame", "4-post-frame", "4-post-cabinet", "wall-frame", "wall-frame-vertical", "wall-cabinet", "wall-cabinet-vertical"}

func resourceNetboxRack() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxRackCreate,
		Read:   resourceNetboxRackRead,
		Update: resourceNetboxRackUpdate,
		Delete: resourceNetboxRackDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/rack/):

> The rack model represents a physical two- or four-post equipment rack in which devices can be installed. Each rack must be assigned to a site, and may optionally be assigned to a location within that site. Racks can also be organized by user-defined functional roles. The name and facility ID of each rack within a location must be unique.

Rack height is measured in rack units (U); racks are commonly between 42U and 48U tall, but NetBox allows you to define racks of arbitrary height. A toggle is provided to indicate whether rack units are in ascending (from the ground up) or descending order.

Each rack is assigned a name and (optionally) a separate facility ID. This is helpful when leasing space in a data center your organization does not own: The facility will often assign a seemingly arbitrary ID to a rack (for example, "M204.313") whereas internally you refer to is simply as "R113." A unique serial number and asset tag may also be associated with each rack.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"site_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxRackStatusOptions, false),
				Description:  buildValidValueDescription(resourceNetboxRackStatusOptions),
			},
			"width": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"rack_type_id"},
				ValidateFunc:  validation.IntInSlice(resourceNetboxRackWidthOptions),
				Description:   "Valid values are `10`, `19`, `21` and `23`. Inherited from `rack_type_id` when set.",
			},
			"u_height": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"rack_type_id"},
				ValidateFunc:  validation.IntBetween(1, 100),
			},
			"rack_type_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"width", "u_height", "form_factor"},
				Description:   "Reference to a `netbox_rack_type`. When set, the rack's physical dimensions (such as `width`, `u_height` and `form_factor`) are inherited from the rack type.",
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"starting_unit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 32767),
				Description:  "Starting unit for rack",
			},
			"outer_height": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 32767),
				Description:  "Outer dimension of rack (height)",
			},
			"airflow": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [front-to-rear, rear-to-front]",
				ValidateFunc: validation.StringInSlice([]string{"front-to-rear", "rear-to-front"}, false),
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
			tagsKey: tagsSchema,
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"facility_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"location_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"role_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"serial": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"asset_tag": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"weight": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"max_weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validatePositiveInt32,
			},
			"weight_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"weight", "max_weight"},
				ValidateFunc: validation.StringInSlice(resourceNetboxRackWeightUnitOptions, false),
				Description:  buildValidValueDescription(resourceNetboxRackWeightUnitOptions),
			},
			"desc_units": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If rack units are descending",
				Default:     false,
			},
			"outer_width": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validatePositiveInt16,
			},
			"outer_depth": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validatePositiveInt16,
			},
			"outer_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"outer_width", "outer_depth"},
				ValidateFunc: validation.StringInSlice(resourceNetboxRackOuterUnitOptions, false),
				Description:  buildValidValueDescription(resourceNetboxRackOuterUnitOptions),
			},
			"mounting_depth": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validatePositiveInt16,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			customFieldsKey: customFieldsSchema,
			"form_factor": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"rack_type_id"},
				ValidateFunc:  validation.StringInSlice(resourceNetboxRackFormFactorOptions, false),
				Description:   buildValidValueDescription(resourceNetboxRackFormFactorOptions),
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const rackAPIPath = "dcim/racks/"

type rackRequest struct {
	Name              string             `json:"name"`
	Site              *int64             `json:"site"`
	Status            string             `json:"status,omitempty"`
	Width             *int64             `json:"width,omitempty"`
	UHeight           int64              `json:"u_height,omitempty"`
	Tenant            *int64             `json:"tenant"`
	FacilityId        *string            `json:"facility_id"`
	Location          *int64             `json:"location"`
	Role              *int64             `json:"role"`
	Serial            string             `json:"serial,omitempty"`
	AssetTag          *string            `json:"asset_tag"`
	Weight            *float64           `json:"weight"`
	MaxWeight         *int64             `json:"max_weight"`
	WeightUnit        string             `json:"weight_unit"`
	DescUnits         bool               `json:"desc_units"`
	OuterWidth        *int64             `json:"outer_width"`
	OuterDepth        *int64             `json:"outer_depth"`
	OuterUnit         string             `json:"outer_unit"`
	MountingDepth     *int64             `json:"mounting_depth"`
	Description       string             `json:"description,omitempty"`
	Comments          string             `json:"comments,omitempty"`
	FormFactor        string             `json:"form_factor"`
	RackType          *int64             `json:"rack_type,omitempty"`
	Owner             *int64             `json:"owner"`
	Group             *int64             `json:"group"`
	StartingUnit      int64              `json:"starting_unit,omitempty"`
	OuterHeight       *int64             `json:"outer_height"`
	Airflow           *string            `json:"airflow,omitempty"`
	CoolingCapability *string            `json:"cooling_capability,omitempty"`
	CoolingCapacity   *float64           `json:"cooling_capacity"`
	CustomFields      interface{}        `json:"custom_fields,omitempty"`
	Tags              []netboxapi.TagRef `json:"tags"`
}

type rackResponse struct {
	Name              string                 `json:"name"`
	Site              *netboxapi.Ref         `json:"site"`
	Status            *netboxapi.Choice      `json:"status"`
	Width             *netboxapi.IntChoice   `json:"width"`
	UHeight           int64                  `json:"u_height"`
	Tenant            *netboxapi.Ref         `json:"tenant"`
	FacilityId        *string                `json:"facility_id"`
	Location          *netboxapi.Ref         `json:"location"`
	Role              *netboxapi.Ref         `json:"role"`
	Serial            string                 `json:"serial"`
	AssetTag          *string                `json:"asset_tag"`
	Weight            *float64               `json:"weight"`
	MaxWeight         *int64                 `json:"max_weight"`
	WeightUnit        *netboxapi.Choice      `json:"weight_unit"`
	DescUnits         bool                   `json:"desc_units"`
	OuterWidth        *int64                 `json:"outer_width"`
	OuterDepth        *int64                 `json:"outer_depth"`
	OuterUnit         *netboxapi.Choice      `json:"outer_unit"`
	MountingDepth     *int64                 `json:"mounting_depth"`
	Description       string                 `json:"description"`
	Comments          string                 `json:"comments"`
	FormFactor        *netboxapi.Choice      `json:"form_factor"`
	RackType          *netboxapi.Ref         `json:"rack_type"`
	Owner             *netboxapi.Ref         `json:"owner"`
	Group             *netboxapi.Ref         `json:"group"`
	StartingUnit      *int64                 `json:"starting_unit"`
	OuterHeight       *int64                 `json:"outer_height"`
	Airflow           *netboxapi.Choice      `json:"airflow"`
	CoolingCapability *netboxapi.Choice      `json:"cooling_capability"`
	CoolingCapacity   *float64               `json:"cooling_capacity"`
	CustomFields      map[string]interface{} `json:"custom_fields"`
	Tags              []netboxapi.TagRef     `json:"tags"`
}

func rackRequestFromResourceData(api *providerState, d *schema.ResourceData) (*rackRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &rackRequest{
		CoolingCapacity:   getOptionalFloat(d, "cooling_capacity"),
		CoolingCapability: getOptionalStrClear(d, "cooling_capability"),
		Airflow:           getOptionalStrClear(d, "airflow"),
		OuterHeight:       getOptionalInt(d, "outer_height"),
		StartingUnit:      int64(d.Get("starting_unit").(int)),
		Group:             getOptionalInt(d, "group_id"),
		Owner:             getOptionalInt(d, "owner_id"),
		Name:              name,
		Site:              int64ToPtr(int64(d.Get("site_id").(int))),
		Status:            d.Get("status").(string),
		Width:             getOptionalInt(d, "width"),
		UHeight:           int64(d.Get("u_height").(int)),
		Tenant:            getOptionalInt(d, "tenant_id"),
		FacilityId:        nullableString(getOptionalStr(d, "facility_id", false)),
		Location:          getOptionalInt(d, "location_id"),
		Role:              getOptionalInt(d, "role_id"),
		Serial:            getOptionalStr(d, "serial", true),
		AssetTag:          nullableString(getOptionalStr(d, "asset_tag", false)),
		Weight:            getOptionalFloat(d, "weight"),
		MaxWeight:         getOptionalInt(d, "max_weight"),
		WeightUnit:        getOptionalStr(d, "weight_unit", false),
		DescUnits:         d.Get("desc_units").(bool),
		OuterWidth:        getOptionalInt(d, "outer_width"),
		OuterDepth:        getOptionalInt(d, "outer_depth"),
		OuterUnit:         getOptionalStr(d, "outer_unit", false),
		MountingDepth:     getOptionalInt(d, "mounting_depth"),
		Description:       getOptionalStr(d, "description", true),
		Comments:          getOptionalStr(d, "comments", true),
		FormFactor:        getOptionalStr(d, "form_factor", false),
		RackType:          getOptionalInt(d, "rack_type_id"),
		Tags:              tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxRackCreate(d *schema.ResourceData, m interface{}) error {
	data, err := rackRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, rackAPIPath, data, resourceNetboxRackRead)
}

func resourceNetboxRackRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, rackAPIPath, func(r *rackResponse) {
		d.Set("cooling_capacity", r.CoolingCapacity)
		d.Set("cooling_capability", netboxapi.ChoiceValue(r.CoolingCapability))
		d.Set("airflow", netboxapi.ChoiceValue(r.Airflow))
		d.Set("outer_height", r.OuterHeight)
		d.Set("starting_unit", r.StartingUnit)
		d.Set("group_id", netboxapi.RefID(r.Group))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("site_id", netboxapi.RefID(r.Site))
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		if r.Width != nil {
			d.Set("width", r.Width.Value)
		} else {
			d.Set("width", nil)
		}
		d.Set("u_height", r.UHeight)
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("facility_id", r.FacilityId)
		d.Set("location_id", netboxapi.RefID(r.Location))
		d.Set("role_id", netboxapi.RefID(r.Role))
		d.Set("serial", r.Serial)
		d.Set("asset_tag", r.AssetTag)
		d.Set("weight", r.Weight)
		d.Set("max_weight", r.MaxWeight)
		d.Set("weight_unit", netboxapi.ChoiceValue(r.WeightUnit))
		d.Set("desc_units", r.DescUnits)
		d.Set("outer_width", r.OuterWidth)
		d.Set("outer_depth", r.OuterDepth)
		d.Set("outer_unit", netboxapi.ChoiceValue(r.OuterUnit))
		d.Set("mounting_depth", r.MountingDepth)
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)
		d.Set("form_factor", netboxapi.ChoiceValue(r.FormFactor))
		d.Set("rack_type_id", netboxapi.RefID(r.RackType))
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxRackUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := rackRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, rackAPIPath, data, resourceNetboxRackRead)
}

func resourceNetboxRackDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, rackAPIPath)
}
