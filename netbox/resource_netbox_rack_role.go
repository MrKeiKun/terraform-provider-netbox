package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxRackRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxRackRoleCreate,
		Read:   resourceNetboxRackRoleRead,
		Update: resourceNetboxRackRoleUpdate,
		Delete: resourceNetboxRackRoleDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/rackrole/):

> Each rack can optionally be assigned a user-defined functional role. For example, you might designate a rack for compute or storage resources, or to house colocated customer devices.`,

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
			"color_hex": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const rackRoleAPIPath = "dcim/rack-roles/"

type rackRoleRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Color        string             `json:"color,omitempty"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type rackRoleResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Color        string                 `json:"color"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func rackRoleRequestFromResourceData(api *providerState, d *schema.ResourceData) *rackRoleRequest {
	name := d.Get("name").(string)
	// A failed tag lookup leaves the tags empty, same as before.
	tags, _ := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	return &rackRoleRequest{
		CustomFields: customFieldsRequest(d),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Color:        d.Get("color_hex").(string),
		Description:  getOptionalStr(d, "description", true),
		Tags:         tags,
	}
}

func resourceNetboxRackRoleCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, rackRoleAPIPath, rackRoleRequestFromResourceData(m.(*providerState), d), resourceNetboxRackRoleRead)
}

func resourceNetboxRackRoleRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, rackRoleAPIPath, func(rackRole *rackRoleResponse) {
		d.Set(customFieldsKey, getCustomFields(rackRole.CustomFields))
		d.Set("comments", rackRole.Comments)
		d.Set("owner_id", netboxapi.RefID(rackRole.Owner))
		d.Set("name", rackRole.Name)
		d.Set("slug", rackRole.Slug)
		d.Set("description", rackRole.Description)
		d.Set("color_hex", rackRole.Color)
		api.readTagRefs(d, rackRole.Tags)
	})
}

func resourceNetboxRackRoleUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, rackRoleAPIPath, rackRoleRequestFromResourceData(m.(*providerState), d), resourceNetboxRackRoleRead)
}

func resourceNetboxRackRoleDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, rackRoleAPIPath)
}
