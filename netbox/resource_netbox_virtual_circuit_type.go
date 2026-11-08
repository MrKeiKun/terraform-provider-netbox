package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxVirtualCircuitType() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVirtualCircuitTypeCreate,
		Read:   resourceNetboxVirtualCircuitTypeRead,
		Update: resourceNetboxVirtualCircuitTypeUpdate,
		Delete: resourceNetboxVirtualCircuitTypeDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/models/circuits/virtualcircuittype/):

> A virtual circuit type classifies virtual circuits.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the virtual circuit type.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the virtual circuit type.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"color_hex": {
				Type:        schema.TypeString,
				Description: "The color hex of the virtual circuit type.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the virtual circuit type.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the virtual circuit type.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the virtual circuit type.",
				Optional:    true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const virtualCircuitTypeAPIPath = "circuits/virtual-circuit-types/"

type virtualCircuitTypeRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Color        string             `json:"color,omitempty"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type virtualCircuitTypeResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Color        string                 `json:"color"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func virtualCircuitTypeRequestFromResourceData(api *providerState, d *schema.ResourceData) (*virtualCircuitTypeRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &virtualCircuitTypeRequest{
		Name:        d.Get("name").(string),
		Slug:        slugFromResourceData(d, name),
		Color:       getOptionalStr(d, "color_hex", true),
		Description: getOptionalStr(d, "description", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxVirtualCircuitTypeCreate(d *schema.ResourceData, m interface{}) error {
	data, err := virtualCircuitTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, virtualCircuitTypeAPIPath, data, resourceNetboxVirtualCircuitTypeRead)
}

func resourceNetboxVirtualCircuitTypeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, virtualCircuitTypeAPIPath, func(r *virtualCircuitTypeResponse) {
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("color_hex", r.Color)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxVirtualCircuitTypeUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := virtualCircuitTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, virtualCircuitTypeAPIPath, data, resourceNetboxVirtualCircuitTypeRead)
}

func resourceNetboxVirtualCircuitTypeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, virtualCircuitTypeAPIPath)
}
