package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCircuitType() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCircuitTypeCreate,
		Read:   resourceNetboxCircuitTypeRead,
		Update: resourceNetboxCircuitTypeUpdate,
		Delete: resourceNetboxCircuitTypeDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/features/circuits/#circuit-types):

> Circuits are classified by functional type. These types are completely customizable, and are typically used to convey the type of service being delivered over a circuit.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the circuit type.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the circuit type.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the circuit type.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the circuit type.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the circuit type.",
				Optional:    true,
			},
			"color_hex": {
				Type:        schema.TypeString,
				Description: "The color hex of the circuit type.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const circuitTypeAPIPath = "circuits/circuit-types/"

type circuitTypeRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Color        string             `json:"color,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type circuitTypeResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Color        string                 `json:"color"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func circuitTypeRequestFromResourceData(d *schema.ResourceData) *circuitTypeRequest {
	name := d.Get("name").(string)
	return &circuitTypeRequest{
		CustomFields: customFieldsRequest(d),
		Color:        getOptionalStr(d, "color_hex", true),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Description:  d.Get("description").(string),
		Tags:         []netboxapi.TagRef{},
	}
}

func resourceNetboxCircuitTypeCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, circuitTypeAPIPath, circuitTypeRequestFromResourceData(d), resourceNetboxCircuitTypeRead)
}

func resourceNetboxCircuitTypeRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, circuitTypeAPIPath, func(res *circuitTypeResponse) {
		d.Set(customFieldsKey, getCustomFields(res.CustomFields))
		d.Set("color_hex", res.Color)
		d.Set("comments", res.Comments)
		d.Set("owner_id", netboxapi.RefID(res.Owner))
		d.Set("name", res.Name)
		d.Set("slug", res.Slug)
		d.Set("description", res.Description)
	})
}

func resourceNetboxCircuitTypeUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, circuitTypeAPIPath, circuitTypeRequestFromResourceData(d), resourceNetboxCircuitTypeRead)
}

func resourceNetboxCircuitTypeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, circuitTypeAPIPath)
}
