package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCable() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCableCreate,
		Read:   resourceNetboxCableRead,
		Update: resourceNetboxCableUpdate,
		Delete: resourceNetboxCableDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/cable/):

> All connections between device components in NetBox are represented using cables. A cable represents a direct physical connection between two sets of endpoints (A and B), such as a console port and a patch panel port, or between two network interfaces.`,

		Schema: map[string]*schema.Schema{
			"a_termination": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     genericObjectSchema,
			},
			"b_termination": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     genericObjectSchema,
			},
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [connected, planned, decommissioning]",
				ValidateFunc: validation.StringInSlice([]string{"connected", "planned", "decommissioning"}, false),
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "One of [cat3, cat5, cat5e, cat6, cat6a, cat7, cat7a, cat8, dac-active, dac-passive, mrj21-trunk, coaxial, mmf, mmf-om1, mmf-om2, mmf-om3, mmf-om4, mmf-om5, smf, smf-os1, smf-os2, aoc, power, usb]",
				ValidateFunc: validation.StringInSlice([]string{
					"cat3", "cat5", "cat5e", "cat6", "cat6a", "cat7", "cat7a", "cat8", "dac-active",
					"dac-passive", "mrj21-trunk", "coaxial", "mmf", "mmf-om1", "mmf-om2", "mmf-om3",
					"mmf-om4", "mmf-om5", "smf", "smf-os1", "smf-os2", "aoc", "power", "usb",
				}, false),
			},
			"tenant_id": {
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
			"length": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"length_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"length"},
				Description:  "One of [km, m, cm, mi, ft, in]",
				ValidateFunc: validation.StringInSlice([]string{"km", "m", "cm", "mi", "ft", "in"}, false),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
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

const cableAPIPath = "dcim/cables/"

type cableRequest struct {
	Status        string             `json:"status,omitempty"`
	Type          string             `json:"type"`
	Tenant        *int64             `json:"tenant"`
	Label         string             `json:"label,omitempty"`
	Color         string             `json:"color"`
	Length        *float64           `json:"length"`
	LengthUnit    string             `json:"length_unit,omitempty"`
	Description   string             `json:"description,omitempty"`
	Comments      string             `json:"comments,omitempty"`
	ATerminations []genericObject    `json:"a_terminations"`
	BTerminations []genericObject    `json:"b_terminations"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

type cableResponse struct {
	Status        *netboxapi.Choice      `json:"status"`
	Type          string                 `json:"type"`
	Tenant        *netboxapi.Ref         `json:"tenant"`
	Label         string                 `json:"label"`
	Color         string                 `json:"color"`
	Length        *float64               `json:"length"`
	LengthUnit    *netboxapi.Choice      `json:"length_unit"`
	Description   string                 `json:"description"`
	Comments      string                 `json:"comments"`
	ATerminations []genericObject        `json:"a_terminations"`
	BTerminations []genericObject        `json:"b_terminations"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
	Tags          []netboxapi.TagRef     `json:"tags"`
}

func cableRequestFromResourceData(api *providerState, d *schema.ResourceData) (*cableRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &cableRequest{
		Status:      d.Get("status").(string),
		Type:        d.Get("type").(string),
		Tenant:      getOptionalInt(d, "tenant_id"),
		Label:       getOptionalStr(d, "label", true),
		Color:       d.Get("color_hex").(string),
		Length:      getOptionalFloat(d, "length"),
		LengthUnit:  d.Get("length_unit").(string),
		Description: getOptionalStr(d, "description", true),
		Comments:    getOptionalStr(d, "comments", true),

		Tags: tags,
	}

	data.ATerminations = getGenericObjectsFromSchemaSet(d.Get("a_termination").(*schema.Set))
	data.BTerminations = getGenericObjectsFromSchemaSet(d.Get("b_termination").(*schema.Set))
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxCableCreate(d *schema.ResourceData, m interface{}) error {
	data, err := cableRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, cableAPIPath, data, resourceNetboxCableRead)
}

func resourceNetboxCableRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, cableAPIPath, func(r *cableResponse) {
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("type", r.Type)
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("label", r.Label)
		d.Set("color_hex", r.Color)
		d.Set("length", r.Length)
		d.Set("length_unit", netboxapi.ChoiceValue(r.LengthUnit))
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)

		d.Set("a_termination", getSchemaSetFromGenericObjects(r.ATerminations))
		d.Set("b_termination", getSchemaSetFromGenericObjects(r.BTerminations))
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxCableUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := cableRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, cableAPIPath, data, resourceNetboxCableRead)
}

func resourceNetboxCableDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, cableAPIPath)
}
