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
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
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

const circuitTypeAPIPath = "circuits/circuit-types/"

type circuitTypeRequest struct {
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type circuitTypeResponse struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func circuitTypeRequestFromResourceData(d *schema.ResourceData) *circuitTypeRequest {
	name := d.Get("name").(string)
	return &circuitTypeRequest{
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Description: d.Get("description").(string),
		Tags:        []netboxapi.TagRef{},
	}
}

func resourceNetboxCircuitTypeCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, circuitTypeAPIPath, circuitTypeRequestFromResourceData(d), resourceNetboxCircuitTypeRead)
}

func resourceNetboxCircuitTypeRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, circuitTypeAPIPath, func(res *circuitTypeResponse) {
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
