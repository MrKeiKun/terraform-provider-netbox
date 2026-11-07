package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDeviceBay() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDeviceBayCreate,
		Read:   resourceNetboxDeviceBayRead,
		Update: resourceNetboxDeviceBayUpdate,
		Delete: resourceNetboxDeviceBayDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/devicebay/):

> Device bays represent a space or slot within a device in which a field-replaceable device may be installed. A common example is that of a chassis-based server that holds a number of blades which may contain device components that don't fit the module pattern. Devices in turn hold additional components that become available to the parent device.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"installed_device_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
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

const deviceBayAPIPath = "dcim/device-bays/"

type deviceBayRequest struct {
	Name            string             `json:"name"`
	Device          *int64             `json:"device"`
	Label           string             `json:"label,omitempty"`
	InstalledDevice *int64             `json:"installed_device,omitempty"`
	Description     string             `json:"description,omitempty"`
	CustomFields    interface{}        `json:"custom_fields,omitempty"`
	Tags            []netboxapi.TagRef `json:"tags"`
}

type deviceBayResponse struct {
	Name            string                 `json:"name"`
	Device          *netboxapi.Ref         `json:"device"`
	Label           string                 `json:"label"`
	InstalledDevice *netboxapi.Ref         `json:"installed_device"`
	Description     string                 `json:"description"`
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
		Name:            name,
		Device:          int64ToPtr(int64(d.Get("device_id").(int))),
		Label:           getOptionalStr(d, "label", true),
		InstalledDevice: getOptionalInt(d, "installed_device_id"),
		Description:     getOptionalStr(d, "description", true),
		Tags:            tags,
	}
	if ct, ok := d.GetOk(customFieldsKey); ok {
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
		d.Set("name", r.Name)
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("label", r.Label)
		if r.InstalledDevice != nil {
			d.Set("installed_device_id", r.InstalledDevice.ID)
		}
		d.Set("description", r.Description)
		if cf := getCustomFields(r.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}
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
