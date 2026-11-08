package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxManufacturer() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxManufacturerCreate,
		Read:   resourceNetboxManufacturerRead,
		Update: resourceNetboxManufacturerUpdate,
		Delete: resourceNetboxManufacturerDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/device-types/#manufacturers):

> A manufacturer represents the "make" of a device; e.g. Cisco or Dell. Each device type must be assigned to a manufacturer. (Inventory items and platforms may also be associated with manufacturers.) Each manufacturer must have a unique name and may have a description assigned to it.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the manufacturer.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the manufacturer.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the manufacturer.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the manufacturer.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the manufacturer.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const manufacturerAPIPath = "dcim/manufacturers/"

type manufacturerRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Description  string             `json:"description,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type manufacturerResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Description  string                 `json:"description"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func manufacturerRequestFromResourceData(d *schema.ResourceData) *manufacturerRequest {
	name := d.Get("name").(string)
	return &manufacturerRequest{
		CustomFields: customFieldsRequest(d),
		Description:  getOptionalStr(d, "description", true),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Tags:         []netboxapi.TagRef{},
	}
}

func resourceNetboxManufacturerCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, manufacturerAPIPath, manufacturerRequestFromResourceData(d), resourceNetboxManufacturerRead)
}

func resourceNetboxManufacturerRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, manufacturerAPIPath, func(res *manufacturerResponse) {
		d.Set(customFieldsKey, getCustomFields(res.CustomFields))
		d.Set("description", res.Description)
		d.Set("comments", res.Comments)
		d.Set("owner_id", netboxapi.RefID(res.Owner))
		d.Set("name", res.Name)
		d.Set("slug", res.Slug)
	})
}

func resourceNetboxManufacturerUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, manufacturerAPIPath, manufacturerRequestFromResourceData(d), resourceNetboxManufacturerRead)
}

func resourceNetboxManufacturerDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, manufacturerAPIPath)
}
