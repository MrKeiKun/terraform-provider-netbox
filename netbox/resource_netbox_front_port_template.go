package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxFrontPortTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxFrontPortTemplateCreate,
		Read:   resourceNetboxFrontPortTemplateRead,
		Update: resourceNetboxFrontPortTemplateUpdate,
		Delete: resourceNetboxFrontPortTemplateDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/frontporttemplate/):

> A front port template describes a front port that is created on every device of a device type or module type.`,

		Schema: map[string]*schema.Schema{
			"device_type_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"module_type_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "{module} is accepted as a substitution for the module bay position when attached to a module type.",
			},
			"label": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Physical label",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [8p8c, 8p6c, 8p4c, 8p2c, 6p6c, 6p4c, 6p2c, 4p4c, 4p2c, gg45, tera-4p, tera-2p, tera-1p, 110-punch, bnc, f, n, mrj21, fc, fc-pc, fc-upc, fc-apc, lc, lc-pc, lc-upc, lc-apc, mu, mu-pc, mu-upc, mu-apc, lsh, lsh-pc, lsh-upc, lsh-apc, lx5, lx5-pc, lx5-upc, lx5-apc, mpo, mtrj, sc, sc-pc, sc-upc, sc-apc, st, cs, sn, mdc, sma-905, sma-906, urm-p2, urm-p4, urm-p8, splice, usb-a, usb-b, usb-c, usb-mini-a, usb-mini-b, usb-micro-a, usb-micro-b, usb-micro-ab, other]",
				ValidateFunc: validation.StringInSlice([]string{"8p8c", "8p6c", "8p4c", "8p2c", "6p6c", "6p4c", "6p2c", "4p4c", "4p2c", "gg45", "tera-4p", "tera-2p", "tera-1p", "110-punch", "bnc", "f", "n", "mrj21", "fc", "fc-pc", "fc-upc", "fc-apc", "lc", "lc-pc", "lc-upc", "lc-apc", "mu", "mu-pc", "mu-upc", "mu-apc", "lsh", "lsh-pc", "lsh-upc", "lsh-apc", "lx5", "lx5-pc", "lx5-upc", "lx5-apc", "mpo", "mtrj", "sc", "sc-pc", "sc-upc", "sc-apc", "st", "cs", "sn", "mdc", "sma-905", "sma-906", "urm-p2", "urm-p4", "urm-p8", "splice", "usb-a", "usb-b", "usb-c", "usb-mini-a", "usb-mini-b", "usb-micro-a", "usb-micro-b", "usb-micro-ab", "other"}, false),
			},
			"color_hex": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"positions": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 1024),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const frontPortTemplateAPIPath = "dcim/front-port-templates/"

type frontPortTemplateRequest struct {
	DeviceType  *int64 `json:"device_type"`
	ModuleType  *int64 `json:"module_type"`
	Name        string `json:"name"`
	Label       string `json:"label,omitempty"`
	Type        string `json:"type,omitempty"`
	Color       string `json:"color,omitempty"`
	Positions   int64  `json:"positions,omitempty"`
	Description string `json:"description,omitempty"`
}

type frontPortTemplateResponse struct {
	DeviceType  *netboxapi.Ref    `json:"device_type"`
	ModuleType  *netboxapi.Ref    `json:"module_type"`
	Name        string            `json:"name"`
	Label       string            `json:"label"`
	Type        *netboxapi.Choice `json:"type"`
	Color       string            `json:"color"`
	Positions   *int64            `json:"positions"`
	Description string            `json:"description"`
}

func frontPortTemplateRequestFromResourceData(api *providerState, d *schema.ResourceData) (*frontPortTemplateRequest, error) {
	data := &frontPortTemplateRequest{
		DeviceType:  getOptionalInt(d, "device_type_id"),
		ModuleType:  getOptionalInt(d, "module_type_id"),
		Name:        d.Get("name").(string),
		Label:       getOptionalStr(d, "label", true),
		Type:        d.Get("type").(string),
		Color:       getOptionalStr(d, "color_hex", true),
		Positions:   int64(d.Get("positions").(int)),
		Description: getOptionalStr(d, "description", true),
	}
	return data, nil
}

func resourceNetboxFrontPortTemplateCreate(d *schema.ResourceData, m interface{}) error {
	data, err := frontPortTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, frontPortTemplateAPIPath, data, resourceNetboxFrontPortTemplateRead)
}

func resourceNetboxFrontPortTemplateRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, frontPortTemplateAPIPath, func(r *frontPortTemplateResponse) {
		d.Set("device_type_id", netboxapi.RefID(r.DeviceType))
		d.Set("module_type_id", netboxapi.RefID(r.ModuleType))
		d.Set("name", r.Name)
		d.Set("label", r.Label)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("color_hex", r.Color)
		d.Set("positions", r.Positions)
		d.Set("description", r.Description)
	})
}

func resourceNetboxFrontPortTemplateUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := frontPortTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, frontPortTemplateAPIPath, data, resourceNetboxFrontPortTemplateRead)
}

func resourceNetboxFrontPortTemplateDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, frontPortTemplateAPIPath)
}
