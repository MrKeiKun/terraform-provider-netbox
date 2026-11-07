package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDeviceFrontPort() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDeviceFrontPortCreate,
		Read:   resourceNetboxDeviceFrontPortRead,
		Update: resourceNetboxDeviceFrontPortUpdate,
		Delete: resourceNetboxDeviceFrontPortDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/frontport/):

> Front ports are pass-through ports which represent physical cable connections that comprise part of a longer path. For example, the ports on the front face of a UTP patch panel would be modeled in NetBox as front ports. Each port is assigned a physical type, and must be mapped to a specific rear port on the same device. A single rear port may be mapped to multiple front ports, using numeric positions to annotate the specific alignment of each.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "One of [8p8c, 8p6c, 8p4c, 8p2c, 6p6c, 6p4c, 6p2c, 4p4c, 4p2c, gg45, tera-4p, tera-2p, tera-1p, 110-punch, bnc, f, n, mrj21, fc, lc, lc-pc, lc-upc, lc-apc, lsh, lsh-pc, lsh-upc, lsh-apc, mpo, mtrj, sc, sc-pc, sc-upc, sc-apc, st, cs, sn, sma-905, sma-906, urm-p2, urm-p4, urm-p8, splice, other]",
			},
			"rear_port_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"rear_port_position": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"module_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"color_hex": {
				Type:     schema.TypeString,
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

const deviceFrontPortAPIPath = "dcim/front-ports/"

type deviceFrontPortRequest struct {
	Name             string             `json:"name"`
	Device           *int64             `json:"device"`
	Module           *int64             `json:"module"`
	Label            string             `json:"label,omitempty"`
	Type             string             `json:"type"`
	RearPort         *int64             `json:"rear_port"`
	RearPortPosition int64              `json:"rear_port_position,omitempty"`
	Color            string             `json:"color"`
	Description      string             `json:"description,omitempty"`
	MarkConnected    bool               `json:"mark_connected"`
	CustomFields     interface{}        `json:"custom_fields,omitempty"`
	Tags             []netboxapi.TagRef `json:"tags"`
}

type deviceFrontPortResponse struct {
	Name             string                 `json:"name"`
	Device           *netboxapi.Ref         `json:"device"`
	Module           *netboxapi.Ref         `json:"module"`
	Label            string                 `json:"label"`
	Type             *netboxapi.Choice      `json:"type"`
	RearPort         *netboxapi.Ref         `json:"rear_port"`
	RearPortPosition int64                  `json:"rear_port_position"`
	Color            string                 `json:"color"`
	Description      string                 `json:"description"`
	MarkConnected    bool                   `json:"mark_connected"`
	CustomFields     map[string]interface{} `json:"custom_fields"`
	Tags             []netboxapi.TagRef     `json:"tags"`
}

func deviceFrontPortRequestFromResourceData(api *providerState, d *schema.ResourceData) (*deviceFrontPortRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &deviceFrontPortRequest{
		Name:             name,
		Device:           int64ToPtr(int64(d.Get("device_id").(int))),
		Module:           getOptionalInt(d, "module_id"),
		Label:            getOptionalStr(d, "label", true),
		Type:             d.Get("type").(string),
		RearPort:         int64ToPtr(int64(d.Get("rear_port_id").(int))),
		RearPortPosition: int64(d.Get("rear_port_position").(int)),
		Color:            getOptionalStr(d, "color_hex", true),
		Description:      getOptionalStr(d, "description", true),
		MarkConnected:    d.Get("mark_connected").(bool),
		Tags:             tags,
	}
	if ct, ok := d.GetOk(customFieldsKey); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxDeviceFrontPortCreate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceFrontPortRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, deviceFrontPortAPIPath, data, resourceNetboxDeviceFrontPortRead)
}

func resourceNetboxDeviceFrontPortRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, deviceFrontPortAPIPath, func(r *deviceFrontPortResponse) {
		d.Set("name", r.Name)
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("module_id", netboxapi.RefID(r.Module))
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("rear_port_id", netboxapi.RefID(r.RearPort))
		d.Set("rear_port_position", r.RearPortPosition)
		d.Set("color_hex", r.Color)
		d.Set("description", r.Description)
		d.Set("mark_connected", r.MarkConnected)
		if cf := getCustomFields(r.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxDeviceFrontPortUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceFrontPortRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, deviceFrontPortAPIPath, data, resourceNetboxDeviceFrontPortRead)
}

func resourceNetboxDeviceFrontPortDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, deviceFrontPortAPIPath)
}
