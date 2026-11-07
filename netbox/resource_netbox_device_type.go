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
	Slug          string             `json:"slug"`
	Model         string             `json:"model"`
	Manufacturer  *int64             `json:"manufacturer"`
	PartNumber    string             `json:"part_number,omitempty"`
	UHeight       *float64           `json:"u_height,omitempty"`
	IsFullDepth   bool               `json:"is_full_depth"`
	SubdeviceRole string             `json:"subdevice_role,omitempty"`
	Description   string             `json:"description,omitempty"`
	Comments      string             `json:"comments,omitempty"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

type deviceTypeResponse struct {
	Slug          string                 `json:"slug"`
	Model         string                 `json:"model"`
	Manufacturer  *netboxapi.Ref         `json:"manufacturer"`
	PartNumber    string                 `json:"part_number"`
	UHeight       *float64               `json:"u_height"`
	IsFullDepth   bool                   `json:"is_full_depth"`
	SubdeviceRole *netboxapi.Choice      `json:"subdevice_role"`
	Description   string                 `json:"description"`
	Comments      string                 `json:"comments"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
	Tags          []netboxapi.TagRef     `json:"tags"`
}

func deviceTypeRequestFromResourceData(api *providerState, d *schema.ResourceData) (*deviceTypeRequest, error) {
	slugSource := d.Get("model").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &deviceTypeRequest{
		Slug:          slugFromResourceData(d, slugSource),
		Model:         d.Get("model").(string),
		PartNumber:    d.Get("part_number").(string),
		UHeight:       float64ToPtr(d.Get("u_height").(float64)),
		IsFullDepth:   d.Get("is_full_depth").(bool),
		SubdeviceRole: d.Get("subdevice_role").(string),
		Description:   d.Get("description").(string),
		Comments:      d.Get("comments").(string),
		Tags:          tags,
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
