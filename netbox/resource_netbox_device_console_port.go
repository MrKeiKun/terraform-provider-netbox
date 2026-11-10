package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDeviceConsolePort() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxDeviceConsolePortCreate),
		ReadContext:   withContext(resourceNetboxDeviceConsolePortRead),
		UpdateContext: withContext(resourceNetboxDeviceConsolePortUpdate),
		DeleteContext: withContext(resourceNetboxDeviceConsolePortDelete),

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/consoleport/):

> A console port provides connectivity to the physical console of a device. These are typically used for temporary access by someone who is physically near the device, or for remote out-of-band access provided via a networked console server.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the device console port.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the device console port.",
				Required:    true,
			},
			"module_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module associated with the device console port.",
				Optional:    true,
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the device console port.",
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
				Description: "A description of the device console port.",
				Optional:    true,
			},
			"mark_connected": {
				Type:        schema.TypeBool,
				Description: "Whether the device console port is marked as connected even without a cable.",
				Default:     false,
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the device console port.",
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

const deviceConsolePortAPIPath = "dcim/console-ports/"

type deviceConsolePortRequest struct {
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

type deviceConsolePortResponse struct {
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

func deviceConsolePortRequestFromResourceData(api *providerState, d *schema.ResourceData) (*deviceConsolePortRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &deviceConsolePortRequest{
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

func resourceNetboxDeviceConsolePortCreate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceConsolePortRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, deviceConsolePortAPIPath, data, resourceNetboxDeviceConsolePortRead)
}

func resourceNetboxDeviceConsolePortRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, deviceConsolePortAPIPath, func(r *deviceConsolePortResponse) {
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

func resourceNetboxDeviceConsolePortUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceConsolePortRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, deviceConsolePortAPIPath, data, resourceNetboxDeviceConsolePortRead)
}

func resourceNetboxDeviceConsolePortDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, deviceConsolePortAPIPath)
}
