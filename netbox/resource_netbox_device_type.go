package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxDeviceType() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDeviceTypeCreate,
		Read:   resourceNetboxDeviceTypeRead,
		Update: resourceNetboxDeviceTypeUpdate,
		Delete: resourceNetboxDeviceTypeDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/device-types/#device-types_1):

> A device type represents a particular make and model of hardware that exists in the real world. Device types define the physical attributes of a device (rack height and depth) and its individual components (console, power, network interfaces, and so on).`,

		Schema: map[string]*schema.Schema{
			"model": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"manufacturer_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"part_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"u_height": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  "1.0",
			},
			"is_full_depth": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"subdevice_role": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"default_platform_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"exclude_from_utilization": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Devices of this type are excluded when calculating rack utilization.",
			},
			"airflow": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [front-to-rear, rear-to-front, left-to-right, right-to-left, side-to-rear, rear-to-side, bottom-to-top, top-to-bottom, passive, mixed]",
				ValidateFunc: validation.StringInSlice([]string{"front-to-rear", "rear-to-front", "left-to-right", "right-to-left", "side-to-rear", "rear-to-side", "bottom-to-top", "top-to-bottom", "passive", "mixed"}, false),
			},
			"cooling_method": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [air, liquid, hybrid, immersion]",
				ValidateFunc: validation.StringInSlice([]string{"air", "liquid", "hybrid", "immersion"}, false),
			},
			"weight": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"weight_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [kg, g, lb, oz]",
				ValidateFunc: validation.StringInSlice([]string{"kg", "g", "lb", "oz"}, false),
			},
			"end_of_life": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The date after which this device type is no longer supported by the manufacturer",
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const deviceTypeAPIPath = "dcim/device-types/"

type deviceTypeRequest struct {
	Slug                   string             `json:"slug"`
	Model                  string             `json:"model"`
	Manufacturer           *int64             `json:"manufacturer"`
	PartNumber             string             `json:"part_number,omitempty"`
	UHeight                *float64           `json:"u_height,omitempty"`
	IsFullDepth            bool               `json:"is_full_depth"`
	SubdeviceRole          string             `json:"subdevice_role,omitempty"`
	Description            string             `json:"description,omitempty"`
	Comments               string             `json:"comments,omitempty"`
	Owner                  *int64             `json:"owner"`
	DefaultPlatform        *int64             `json:"default_platform"`
	ExcludeFromUtilization *bool              `json:"exclude_from_utilization,omitempty"`
	Airflow                *string            `json:"airflow,omitempty"`
	CoolingMethod          *string            `json:"cooling_method,omitempty"`
	Weight                 *float64           `json:"weight"`
	WeightUnit             *string            `json:"weight_unit,omitempty"`
	EndOfLife              string             `json:"end_of_life,omitempty"`
	CustomFields           interface{}        `json:"custom_fields,omitempty"`
	Tags                   []netboxapi.TagRef `json:"tags"`
}

type deviceTypeResponse struct {
	Slug                   string                 `json:"slug"`
	Model                  string                 `json:"model"`
	Manufacturer           *netboxapi.Ref         `json:"manufacturer"`
	PartNumber             string                 `json:"part_number"`
	UHeight                *float64               `json:"u_height"`
	IsFullDepth            bool                   `json:"is_full_depth"`
	SubdeviceRole          *netboxapi.Choice      `json:"subdevice_role"`
	Description            string                 `json:"description"`
	Comments               string                 `json:"comments"`
	Owner                  *netboxapi.Ref         `json:"owner"`
	DefaultPlatform        *netboxapi.Ref         `json:"default_platform"`
	ExcludeFromUtilization bool                   `json:"exclude_from_utilization"`
	Airflow                *netboxapi.Choice      `json:"airflow"`
	CoolingMethod          *netboxapi.Choice      `json:"cooling_method"`
	Weight                 *float64               `json:"weight"`
	WeightUnit             *netboxapi.Choice      `json:"weight_unit"`
	EndOfLife              string                 `json:"end_of_life"`
	CustomFields           map[string]interface{} `json:"custom_fields"`
	Tags                   []netboxapi.TagRef     `json:"tags"`
}

func deviceTypeRequestFromResourceData(api *providerState, d *schema.ResourceData) (*deviceTypeRequest, error) {
	slugSource := d.Get("model").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &deviceTypeRequest{
		EndOfLife:              getOptionalStr(d, "end_of_life", true),
		WeightUnit:             getOptionalStrClear(d, "weight_unit"),
		Weight:                 getOptionalFloat(d, "weight"),
		CoolingMethod:          getOptionalStrClear(d, "cooling_method"),
		Airflow:                getOptionalStrClear(d, "airflow"),
		ExcludeFromUtilization: getOptionalBool(d, "exclude_from_utilization"),
		DefaultPlatform:        getOptionalInt(d, "default_platform_id"),
		Owner:                  getOptionalInt(d, "owner_id"),
		Slug:                   slugFromResourceData(d, slugSource),
		Model:                  d.Get("model").(string),
		PartNumber:             d.Get("part_number").(string),
		UHeight:                float64ToPtr(d.Get("u_height").(float64)),
		IsFullDepth:            d.Get("is_full_depth").(bool),
		SubdeviceRole:          d.Get("subdevice_role").(string),
		Description:            d.Get("description").(string),
		Comments:               d.Get("comments").(string),
		Tags:                   tags,
	}
	if v, ok := d.GetOk("manufacturer_id"); ok {
		data.Manufacturer = int64ToPtr(int64(v.(int)))
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxDeviceTypeCreate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, deviceTypeAPIPath, data, resourceNetboxDeviceTypeRead)
}

func resourceNetboxDeviceTypeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, deviceTypeAPIPath, func(r *deviceTypeResponse) {
		d.Set("end_of_life", r.EndOfLife)
		d.Set("weight_unit", netboxapi.ChoiceValue(r.WeightUnit))
		d.Set("weight", r.Weight)
		d.Set("cooling_method", netboxapi.ChoiceValue(r.CoolingMethod))
		d.Set("airflow", netboxapi.ChoiceValue(r.Airflow))
		d.Set("exclude_from_utilization", r.ExcludeFromUtilization)
		d.Set("default_platform_id", netboxapi.RefID(r.DefaultPlatform))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("slug", r.Slug)
		d.Set("model", r.Model)
		d.Set("manufacturer_id", netboxapi.RefID(r.Manufacturer))
		d.Set("part_number", r.PartNumber)
		d.Set("u_height", r.UHeight)
		d.Set("is_full_depth", r.IsFullDepth)
		d.Set("subdevice_role", netboxapi.ChoiceValue(r.SubdeviceRole))
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxDeviceTypeUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, deviceTypeAPIPath, data, resourceNetboxDeviceTypeRead)
}

func resourceNetboxDeviceTypeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, deviceTypeAPIPath)
}
