package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDevicePowerPort() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDevicePowerPortCreate,
		Read:   resourceNetboxDevicePowerPortRead,
		Update: resourceNetboxDevicePowerPortUpdate,
		Delete: resourceNetboxDevicePowerPortDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/powerport/):

> A power port is a device component which draws power from some external source (e.g. an upstream power outlet), and generally represents a power supply internal to a device.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"module_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "One of [iec-60320-c6, iec-60320-c8, iec-60320-c14, iec-60320-c16, iec-60320-c20, iec-60320-c22, iec-60309-p-n-e-4h, iec-60309-p-n-e-6h, iec-60309-p-n-e-9h, iec-60309-2p-e-4h, iec-60309-2p-e-6h, iec-60309-2p-e-9h, iec-60309-3p-e-4h, iec-60309-3p-e-6h, iec-60309-3p-e-9h, iec-60309-3p-n-e-4h, iec-60309-3p-n-e-6h, iec-60309-3p-n-e-9h, nema-1-15p, nema-5-15p, nema-5-20p, nema-5-30p, nema-5-50p, nema-6-15p, nema-6-20p, nema-6-30p, nema-6-50p, nema-10-30p, nema-10-50p, nema-14-20p, nema-14-30p, nema-14-50p, nema-14-60p, nema-15-15p, nema-15-20p, nema-15-30p, nema-15-50p, nema-15-60p, nema-l1-15p, nema-l5-15p, nema-l5-20p, nema-l5-30p, nema-l5-50p, nema-l6-15p, nema-l6-20p, nema-l6-30p, nema-l6-50p, nema-l10-30p, nema-l14-20p, nema-l14-30p, nema-l14-50p, nema-l14-60p, nema-l15-20p, nema-l15-30p, nema-l15-50p, nema-l15-60p, nema-l21-20p, nema-l21-30p, nema-l22-30p, cs6361c, cs6365c, cs8165c, cs8265c, cs8365c, cs8465c, ita-c, ita-e, ita-f, ita-ef, ita-g, ita-h, ita-i, ita-j, ita-k, ita-l, ita-m, ita-n, ita-o, usb-a, usb-b, usb-c, usb-mini-a, usb-mini-b, usb-micro-a, usb-micro-b, usb-micro-ab, usb-3-b, usb-3-micro-b, dc-terminal, saf-d-grid, neutrik-powercon-20, neutrik-powercon-32, neutrik-powercon-true1, neutrik-powercon-true1-top, ubiquiti-smartpower, hardwired, other]",
			},
			"maximum_draw": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"allocated_draw": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"mark_connected": {
				Type:     schema.TypeBool,
				Default:  false,
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

const devicePowerPortAPIPath = "dcim/power-ports/"

type devicePowerPortRequest struct {
	Name          string             `json:"name"`
	Device        *int64             `json:"device"`
	Module        *int64             `json:"module"`
	Label         string             `json:"label,omitempty"`
	Type          string             `json:"type"`
	MaximumDraw   *int64             `json:"maximum_draw"`
	AllocatedDraw *int64             `json:"allocated_draw"`
	Description   string             `json:"description,omitempty"`
	MarkConnected bool               `json:"mark_connected"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

type devicePowerPortResponse struct {
	Name          string                 `json:"name"`
	Device        *netboxapi.Ref         `json:"device"`
	Module        *netboxapi.Ref         `json:"module"`
	Label         string                 `json:"label"`
	Type          *netboxapi.Choice      `json:"type"`
	MaximumDraw   *int64                 `json:"maximum_draw"`
	AllocatedDraw *int64                 `json:"allocated_draw"`
	Description   string                 `json:"description"`
	MarkConnected bool                   `json:"mark_connected"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
	Tags          []netboxapi.TagRef     `json:"tags"`
}

func devicePowerPortRequestFromResourceData(api *providerState, d *schema.ResourceData) (*devicePowerPortRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &devicePowerPortRequest{
		Name:          name,
		Device:        int64ToPtr(int64(d.Get("device_id").(int))),
		Module:        getOptionalInt(d, "module_id"),
		Label:         getOptionalStr(d, "label", true),
		Type:          getOptionalStr(d, "type", false),
		MaximumDraw:   getOptionalInt(d, "maximum_draw"),
		AllocatedDraw: getOptionalInt(d, "allocated_draw"),
		Description:   getOptionalStr(d, "description", true),
		MarkConnected: d.Get("mark_connected").(bool),
		Tags:          tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxDevicePowerPortCreate(d *schema.ResourceData, m interface{}) error {
	data, err := devicePowerPortRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, devicePowerPortAPIPath, data, resourceNetboxDevicePowerPortRead)
}

func resourceNetboxDevicePowerPortRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, devicePowerPortAPIPath, func(r *devicePowerPortResponse) {
		d.Set("name", r.Name)
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("module_id", netboxapi.RefID(r.Module))
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("maximum_draw", r.MaximumDraw)
		d.Set("allocated_draw", r.AllocatedDraw)
		d.Set("description", r.Description)
		d.Set("mark_connected", r.MarkConnected)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxDevicePowerPortUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := devicePowerPortRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, devicePowerPortAPIPath, data, resourceNetboxDevicePowerPortRead)
}

func resourceNetboxDevicePowerPortDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, devicePowerPortAPIPath)
}
