package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDeviceBay() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxDeviceBayCreate),
		ReadContext:   withContext(resourceNetboxDeviceBayRead),
		UpdateContext: withContext(resourceNetboxDeviceBayUpdate),
		DeleteContext: withContext(resourceNetboxDeviceBayDelete),

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/devicebay/):

> Device bays represent a space or slot within a device in which a field-replaceable device may be installed. A common example is that of a chassis-based server that holds a number of blades which may contain device components that don't fit the module pattern. Devices in turn hold additional components that become available to the parent device.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the device bay.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the device bay.",
				Required:    true,
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the device bay.",
				Optional:    true,
			},
			"installed_device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the installed device associated with the device bay.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the device bay.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the device bay.",
				Optional:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the device bay is enabled.",
				Optional:    true,
				Computed:    true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const deviceBayAPIPath = "dcim/device-bays/"

type deviceBayRequest struct {
	Name            string             `json:"name"`
	Device          *int64             `json:"device"`
	Label           string             `json:"label,omitempty"`
	InstalledDevice *int64             `json:"installed_device,omitempty"`
	Description     string             `json:"description,omitempty"`
	Owner           *int64             `json:"owner"`
	Enabled         *bool              `json:"enabled,omitempty"`
	CustomFields    interface{}        `json:"custom_fields,omitempty"`
	Tags            []netboxapi.TagRef `json:"tags"`
}

type deviceBayResponse struct {
	Name            string                 `json:"name"`
	Device          *netboxapi.Ref         `json:"device"`
	Label           string                 `json:"label"`
	InstalledDevice *netboxapi.Ref         `json:"installed_device"`
	Description     string                 `json:"description"`
	Owner           *netboxapi.Ref         `json:"owner"`
	Enabled         bool                   `json:"enabled"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
	Tags            []netboxapi.TagRef     `json:"tags"`
}

func deviceBayRequestFromResourceData(api *providerState, d *schema.ResourceData) (*deviceBayRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &deviceBayRequest{
		Enabled:         getOptionalBool(d, "enabled"),
		Owner:           getOptionalInt(d, "owner_id"),
		Name:            name,
		Device:          int64ToPtr(int64(d.Get("device_id").(int))),
		Label:           getOptionalStr(d, "label", true),
		InstalledDevice: getOptionalInt(d, "installed_device_id"),
		Description:     getOptionalStr(d, "description", true),
		Tags:            tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxDeviceBayCreate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceBayRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, deviceBayAPIPath, data, resourceNetboxDeviceBayRead)
}

func resourceNetboxDeviceBayRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, deviceBayAPIPath, func(r *deviceBayResponse) {
		d.Set("enabled", r.Enabled)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("label", r.Label)
		if r.InstalledDevice != nil {
			d.Set("installed_device_id", r.InstalledDevice.ID)
		}
		d.Set("description", r.Description)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxDeviceBayUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceBayRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, deviceBayAPIPath, data, resourceNetboxDeviceBayRead)
}

func resourceNetboxDeviceBayDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, deviceBayAPIPath)
}
