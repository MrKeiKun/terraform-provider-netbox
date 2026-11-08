package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxContactRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxContactRoleCreate,
		Read:   resourceNetboxContactRoleRead,
		Update: resourceNetboxContactRoleUpdate,
		Delete: resourceNetboxContactRoleDelete,

		Description: `:meta:subcategory:Tenancy:From the [official documentation](https://docs.netbox.dev/en/stable/features/contacts/#contactroles):

> A contact role defines the relationship of a contact to an assigned object. For example, you might define roles for administrative, operational, and emergency contacts`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the contact role.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the contact role.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the contact role.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the contact role.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the contact role.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const contactRoleAPIPath = "tenancy/contact-roles/"

type contactRoleRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Description  string             `json:"description,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type contactRoleResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Description  string                 `json:"description"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func contactRoleRequestFromResourceData(d *schema.ResourceData) *contactRoleRequest {
	name := d.Get("name").(string)
	return &contactRoleRequest{
		CustomFields: customFieldsRequest(d),
		Description:  getOptionalStr(d, "description", true),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Tags:         []netboxapi.TagRef{},
	}
}

func resourceNetboxContactRoleCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, contactRoleAPIPath, contactRoleRequestFromResourceData(d), resourceNetboxContactRoleRead)
}

func resourceNetboxContactRoleRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, contactRoleAPIPath, func(res *contactRoleResponse) {
		d.Set(customFieldsKey, getCustomFields(res.CustomFields))
		d.Set("description", res.Description)
		d.Set("comments", res.Comments)
		d.Set("owner_id", netboxapi.RefID(res.Owner))
		d.Set("name", res.Name)
		d.Set("slug", res.Slug)
	})
}

func resourceNetboxContactRoleUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, contactRoleAPIPath, contactRoleRequestFromResourceData(d), resourceNetboxContactRoleRead)
}

func resourceNetboxContactRoleDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, contactRoleAPIPath)
}
