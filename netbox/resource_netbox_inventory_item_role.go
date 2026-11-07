package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxInventoryItemRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxInventoryItemRoleCreate,
		Read:   resourceNetboxInventoryItemRoleRead,
		Update: resourceNetboxInventoryItemRoleUpdate,
		Delete: resourceNetboxInventoryItemRoleDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/inventoryitemrole/):

> Inventory items can be organized by functional roles, which are fully customizable by the user. For example, you might create roles for power supplies, fans, interface optics, etc.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"color_hex": {
				Type:     schema.TypeString,
				Required: true,
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

const inventoryItemRoleAPIPath = "dcim/inventory-item-roles/"

type inventoryItemRoleRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Color        string             `json:"color"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type inventoryItemRoleResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Color        string                 `json:"color"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func inventoryItemRoleRequestFromResourceData(api *providerState, d *schema.ResourceData) (*inventoryItemRoleRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &inventoryItemRoleRequest{
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Description: getOptionalStr(d, "description", true),
		Color:       d.Get("color_hex").(string),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxInventoryItemRoleCreate(d *schema.ResourceData, m interface{}) error {
	data, err := inventoryItemRoleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, inventoryItemRoleAPIPath, data, resourceNetboxInventoryItemRoleRead)
}

func resourceNetboxInventoryItemRoleRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, inventoryItemRoleAPIPath, func(r *inventoryItemRoleResponse) {
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
		d.Set("color_hex", r.Color)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxInventoryItemRoleUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := inventoryItemRoleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, inventoryItemRoleAPIPath, data, resourceNetboxInventoryItemRoleRead)
}

func resourceNetboxInventoryItemRoleDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, inventoryItemRoleAPIPath)
}
