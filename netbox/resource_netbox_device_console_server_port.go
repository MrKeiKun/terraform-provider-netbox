package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDeviceConsoleServerPort() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxDeviceConsoleServerPortCreate),
		ReadContext:   withContext(resourceNetboxDeviceConsoleServerPortRead),
		UpdateContext: withContext(resourceNetboxDeviceConsoleServerPortUpdate),
		DeleteContext: withContext(resourceNetboxDeviceConsoleServerPortDelete),

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/consoleserverport/):

> A console server is a device which provides remote access to the local consoles of connected devices. They are typically used to provide remote out-of-band access to network devices, and generally connect to console ports.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the device console server port.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the device console server port.",
				Required:    true,
			},
			"module_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module associated with the device console server port.",
				Optional:    true,
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the device console server port.",
				Optional:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "One of [de-9, db-25, rj-11, rj-12, rj-45, mini-din-8, usb-a, usb-b, usb-c, usb-mini-a, usb-mini-b, usb-micro-a, usb-micro-b, usb-micro-ab, other]",
			},
			"speed": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "One of [1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200]",
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the device console server port.",
				Optional:    true,
			},
			"mark_connected": {
				Type:        schema.TypeBool,
				Description: "Whether the device console server port is marked as connected even without a cable.",
				Default:     false,
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the device console server port.",
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

const deviceConsoleServerPortAPIPath = "dcim/console-server-ports/"

type deviceConsoleServerPortRequest struct {
	Name          string             `json:"name"`
	Device        *int64             `json:"device"`
	Module        *int64             `json:"module"`
	Label         string             `json:"label,omitempty"`
	Type          string             `json:"type"`
	Speed         *int64             `json:"speed"`
	Description   string             `json:"description,omitempty"`
	MarkConnected bool               `json:"mark_connected"`
	Owner         *int64             `json:"owner"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

type deviceConsoleServerPortResponse struct {
	Name          string                 `json:"name"`
	Device        *netboxapi.Ref         `json:"device"`
	Module        *netboxapi.Ref         `json:"module"`
	Label         string                 `json:"label"`
	Type          *netboxapi.Choice      `json:"type"`
	Speed         *netboxapi.IntChoice   `json:"speed"`
	Description   string                 `json:"description"`
	MarkConnected bool                   `json:"mark_connected"`
	Owner         *netboxapi.Ref         `json:"owner"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
	Tags          []netboxapi.TagRef     `json:"tags"`
}

func deviceConsoleServerPortRequestFromResourceData(api *providerState, d *schema.ResourceData) (*deviceConsoleServerPortRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &deviceConsoleServerPortRequest{
		Owner:         getOptionalInt(d, "owner_id"),
		Name:          name,
		Device:        int64ToPtr(int64(d.Get("device_id").(int))),
		Module:        getOptionalInt(d, "module_id"),
		Label:         getOptionalStr(d, "label", true),
		Type:          getOptionalStr(d, "type", false),
		Speed:         getOptionalInt(d, "speed"),
		Description:   getOptionalStr(d, "description", true),
		MarkConnected: d.Get("mark_connected").(bool),
		Tags:          tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxDeviceConsoleServerPortCreate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceConsoleServerPortRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, deviceConsoleServerPortAPIPath, data, resourceNetboxDeviceConsoleServerPortRead)
}

func resourceNetboxDeviceConsoleServerPortRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, deviceConsoleServerPortAPIPath, func(r *deviceConsoleServerPortResponse) {
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("module_id", netboxapi.RefID(r.Module))
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		if r.Speed != nil {
			d.Set("speed", r.Speed.Value)
		} else {
			d.Set("speed", nil)
		}
		d.Set("description", r.Description)
		d.Set("mark_connected", r.MarkConnected)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxDeviceConsoleServerPortUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceConsoleServerPortRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, deviceConsoleServerPortAPIPath, data, resourceNetboxDeviceConsoleServerPortRead)
}

func resourceNetboxDeviceConsoleServerPortDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, deviceConsoleServerPortAPIPath)
}
