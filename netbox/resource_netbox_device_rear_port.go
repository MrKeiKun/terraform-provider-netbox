package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDeviceRearPort() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDeviceRearPortCreate,
		Read:   resourceNetboxDeviceRearPortRead,
		Update: resourceNetboxDeviceRearPortUpdate,
		Delete: resourceNetboxDeviceRearPortDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/rearport/):

> Like front ports, rear ports are pass-through ports which represent the continuation of a path from one cable to the next. Each rear port is defined with its physical type and a number of positions: Rear ports with more than one position can be mapped to multiple front ports. This can be useful for modeling instances where multiple paths share a common cable (for example, six discrete two-strand fiber connections sharing a 12-strand MPO cable).`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the device rear port.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the device rear port.",
				Required:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "One of [8p8c, 8p6c, 8p4c, 8p2c, 6p6c, 6p4c, 6p2c, 4p4c, 4p2c, gg45, tera-4p, tera-2p, tera-1p, 110-punch, bnc, f, n, mrj21, fc, lc, lc-pc, lc-upc, lc-apc, lsh, lsh-pc, lsh-upc, lsh-apc, mpo, mtrj, sc, sc-pc, sc-upc, sc-apc, st, cs, sn, sma-905, sma-906, urm-p2, urm-p4, urm-p8, splice, other]",
			},
			"positions": {
				Type:        schema.TypeInt,
				Description: "The positions of the device rear port.",
				Required:    true,
			},
			"module_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the module associated with the device rear port.",
				Optional:    true,
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the device rear port.",
				Optional:    true,
			},
			"color_hex": {
				Type:        schema.TypeString,
				Description: "The color hex of the device rear port.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the device rear port.",
				Optional:    true,
			},
			"mark_connected": {
				Type:        schema.TypeBool,
				Description: "Whether the device rear port is marked as connected even without a cable.",
				Default:     false,
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the device rear port.",
				Optional:    true,
			},
			"front_ports":   portMappingsSchema("front_port"),
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const deviceRearPortAPIPath = "dcim/rear-ports/"

type deviceRearPortRequest struct {
	FrontPorts    []map[string]int64 `json:"front_ports,omitempty"`
	Name          string             `json:"name"`
	Device        *int64             `json:"device"`
	Module        *int64             `json:"module"`
	Label         string             `json:"label,omitempty"`
	Type          string             `json:"type"`
	Positions     int64              `json:"positions,omitempty"`
	Color         string             `json:"color"`
	Description   string             `json:"description,omitempty"`
	MarkConnected bool               `json:"mark_connected"`
	Owner         *int64             `json:"owner"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

type deviceRearPortResponse struct {
	FrontPorts    []map[string]interface{} `json:"front_ports"`
	Name          string                   `json:"name"`
	Device        *netboxapi.Ref           `json:"device"`
	Module        *netboxapi.Ref           `json:"module"`
	Label         string                   `json:"label"`
	Type          *netboxapi.Choice        `json:"type"`
	Positions     int64                    `json:"positions"`
	Color         string                   `json:"color"`
	Description   string                   `json:"description"`
	MarkConnected bool                     `json:"mark_connected"`
	Owner         *netboxapi.Ref           `json:"owner"`
	CustomFields  map[string]interface{}   `json:"custom_fields"`
	Tags          []netboxapi.TagRef       `json:"tags"`
}

func deviceRearPortRequestFromResourceData(api *providerState, d *schema.ResourceData) (*deviceRearPortRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &deviceRearPortRequest{
		FrontPorts:    portMappingsRequest(d, "front_ports", "front_port"),
		Owner:         getOptionalInt(d, "owner_id"),
		Name:          name,
		Device:        int64ToPtr(int64(d.Get("device_id").(int))),
		Module:        getOptionalInt(d, "module_id"),
		Label:         getOptionalStr(d, "label", true),
		Type:          d.Get("type").(string),
		Positions:     int64(d.Get("positions").(int)),
		Color:         getOptionalStr(d, "color_hex", true),
		Description:   getOptionalStr(d, "description", true),
		MarkConnected: d.Get("mark_connected").(bool),
		Tags:          tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxDeviceRearPortCreate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceRearPortRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, deviceRearPortAPIPath, data, resourceNetboxDeviceRearPortRead)
}

func resourceNetboxDeviceRearPortRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, deviceRearPortAPIPath, func(r *deviceRearPortResponse) {
		d.Set("front_ports", portMappingsState(r.FrontPorts, "front_port"))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("module_id", netboxapi.RefID(r.Module))
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("positions", r.Positions)
		d.Set("color_hex", r.Color)
		d.Set("description", r.Description)
		d.Set("mark_connected", r.MarkConnected)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxDeviceRearPortUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceRearPortRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, deviceRearPortAPIPath, data, resourceNetboxDeviceRearPortRead)
}

func resourceNetboxDeviceRearPortDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, deviceRearPortAPIPath)
}
