package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxDeviceModuleBay() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDeviceModuleBayCreate,
		Read:   resourceNetboxDeviceModuleBayRead,
		Update: resourceNetboxDeviceModuleBayUpdate,
		Delete: resourceNetboxDeviceModuleBayDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/modulebay/):

> Module bays represent a space or slot within a device in which a field-replaceable module may be installed. A common example is that of a chassis-based switch such as the Cisco Nexus 9000 or Juniper EX9200. Modules in turn hold additional components that become available to the parent device.`,

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
			"position": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 30),
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

const deviceModuleBayAPIPath = "dcim/module-bays/"

type deviceModuleBayRequest struct {
	Name         string             `json:"name"`
	Device       *int64             `json:"device"`
	Label        string             `json:"label,omitempty"`
	Position     string             `json:"position,omitempty"`
	Description  string             `json:"description,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type deviceModuleBayResponse struct {
	Name         string                 `json:"name"`
	Device       *netboxapi.Ref         `json:"device"`
	Label        string                 `json:"label"`
	Position     string                 `json:"position"`
	Description  string                 `json:"description"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func deviceModuleBayRequestFromResourceData(api *providerState, d *schema.ResourceData) (*deviceModuleBayRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &deviceModuleBayRequest{
		Name:        name,
		Device:      int64ToPtr(int64(d.Get("device_id").(int))),
		Label:       getOptionalStr(d, "label", true),
		Position:    getOptionalStr(d, "position", true),
		Description: getOptionalStr(d, "description", true),
		Tags:        tags,
	}
	if ct, ok := d.GetOk(customFieldsKey); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxDeviceModuleBayCreate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceModuleBayRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, deviceModuleBayAPIPath, data, resourceNetboxDeviceModuleBayRead)
}

func resourceNetboxDeviceModuleBayRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, deviceModuleBayAPIPath, func(r *deviceModuleBayResponse) {
		d.Set("name", r.Name)
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("label", r.Label)
		d.Set("position", r.Position)
		d.Set("description", r.Description)
		if cf := getCustomFields(r.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxDeviceModuleBayUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceModuleBayRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, deviceModuleBayAPIPath, data, resourceNetboxDeviceModuleBayRead)
}

func resourceNetboxDeviceModuleBayDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, deviceModuleBayAPIPath)
}
