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
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const manufacturerAPIPath = "dcim/manufacturers/"

type manufacturerRequest struct {
	Name string             `json:"name"`
	Slug string             `json:"slug"`
	Tags []netboxapi.TagRef `json:"tags"`
}

type manufacturerResponse struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func manufacturerRequestFromResourceData(d *schema.ResourceData) *manufacturerRequest {
	name := d.Get("name").(string)
	return &manufacturerRequest{
		Name: name,
		Slug: slugFromResourceData(d, name),
		Tags: []netboxapi.TagRef{},
	}
}

func resourceNetboxManufacturerCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, manufacturerAPIPath, manufacturerRequestFromResourceData(d), resourceNetboxManufacturerRead)
}

func resourceNetboxManufacturerRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, manufacturerAPIPath, func(res *manufacturerResponse) {
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
