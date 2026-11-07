package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxModuleBayType() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxModuleBayTypeCreate,
		Read:   resourceNetboxModuleBayTypeRead,
		Update: resourceNetboxModuleBayTypeUpdate,
		Delete: resourceNetboxModuleBayTypeDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/modulebaytype/):

> A module bay type describes the kind of module a module bay accepts, so that modules only fit bays of a matching type.

Requires NetBox v4.7.0 or newer.`,

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
			"manufacturer_id": {
				Type:     schema.TypeInt,
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
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
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

const moduleBayTypeAPIPath = "dcim/module-bay-types/"

type moduleBayTypeRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Manufacturer *int64             `json:"manufacturer"`
	Color        string             `json:"color,omitempty"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type moduleBayTypeResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Manufacturer *netboxapi.Ref         `json:"manufacturer"`
	Color        string                 `json:"color"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func moduleBayTypeRequestFromResourceData(api *providerState, d *schema.ResourceData) (*moduleBayTypeRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &moduleBayTypeRequest{
		Name:         d.Get("name").(string),
		Slug:         slugFromResourceData(d, name),
		Manufacturer: getOptionalInt(d, "manufacturer_id"),
		Color:        getOptionalStr(d, "color_hex", true),
		Description:  getOptionalStr(d, "description", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Comments:     getOptionalStr(d, "comments", true),
		Tags:         tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxModuleBayTypeCreate(d *schema.ResourceData, m interface{}) error {
	data, err := moduleBayTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, moduleBayTypeAPIPath, data, resourceNetboxModuleBayTypeRead)
}

func resourceNetboxModuleBayTypeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, moduleBayTypeAPIPath, func(r *moduleBayTypeResponse) {
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("manufacturer_id", netboxapi.RefID(r.Manufacturer))
		d.Set("color_hex", r.Color)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxModuleBayTypeUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := moduleBayTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, moduleBayTypeAPIPath, data, resourceNetboxModuleBayTypeRead)
}

func resourceNetboxModuleBayTypeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, moduleBayTypeAPIPath)
}
