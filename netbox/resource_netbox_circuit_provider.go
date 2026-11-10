package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCircuitProvider() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxCircuitProviderCreate),
		ReadContext:   withContext(resourceNetboxCircuitProviderRead),
		UpdateContext: withContext(resourceNetboxCircuitProviderUpdate),
		DeleteContext: withContext(resourceNetboxCircuitProviderDelete),

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/features/circuits/#providers):

> A circuit provider is any entity which provides some form of connectivity of among sites or organizations within a site. While this obviously includes carriers which offer Internet and private transit service, it might also include Internet exchange (IX) points and even organizations with whom you peer directly. Each circuit within NetBox must be assigned a provider and a circuit ID which is unique to that provider.
>
> Each provider may be assigned an autonomous system number (ASN), an account number, and contact information.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the circuit provider.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the circuit provider.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the circuit provider.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the circuit provider.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the circuit provider.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const circuitProviderAPIPath = "circuits/providers/"

type circuitProviderRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Asns         []int64            `json:"asns"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type circuitProviderResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func circuitProviderRequestFromResourceData(d *schema.ResourceData) *circuitProviderRequest {
	name := d.Get("name").(string)
	data := &circuitProviderRequest{
		CustomFields: customFieldsRequest(d),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Description:  d.Get("description").(string),
		Asns:         []int64{},
		Tags:         []netboxapi.TagRef{},
	}
	return data
}

func resourceNetboxCircuitProviderCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, circuitProviderAPIPath, circuitProviderRequestFromResourceData(d), resourceNetboxCircuitProviderRead)
}

func resourceNetboxCircuitProviderRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, circuitProviderAPIPath, func(r *circuitProviderResponse) {
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
	})
}

func resourceNetboxCircuitProviderUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, circuitProviderAPIPath, circuitProviderRequestFromResourceData(d), resourceNetboxCircuitProviderRead)
}

func resourceNetboxCircuitProviderDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, circuitProviderAPIPath)
}
