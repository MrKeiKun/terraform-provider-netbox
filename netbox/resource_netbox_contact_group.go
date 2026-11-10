package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxContactGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxContactGroupCreate),
		ReadContext:   withContext(resourceNetboxContactGroupRead),
		UpdateContext: withContext(resourceNetboxContactGroupUpdate),
		DeleteContext: withContext(resourceNetboxContactGroupDelete),

		Description: `:meta:subcategory:Tenancy:From the [official documentation](https://docs.netbox.dev/en/stable/features/contacts/#contact-groups):

> Contacts can be grouped arbitrarily into a recursive hierarchy, and a contact can be assigned to a group at any level within the hierarchy.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the contact group.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the contact group.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"parent_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the parent associated with the contact group.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the contact group.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the contact group.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the contact group.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const contactGroupAPIPath = "tenancy/contact-groups/"

type contactGroupRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Parent       *int64             `json:"parent,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type contactGroupResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Parent       *netboxapi.Ref         `json:"parent"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func contactGroupRequestFromResourceData(d *schema.ResourceData) *contactGroupRequest {
	name := d.Get("name").(string)
	data := &contactGroupRequest{
		CustomFields: customFieldsRequest(d),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Description:  d.Get("description").(string),
		Tags:         []netboxapi.TagRef{},
	}
	if parentID := int64(d.Get("parent_id").(int)); parentID != 0 {
		data.Parent = &parentID
	}
	return data
}

func resourceNetboxContactGroupCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, contactGroupAPIPath, contactGroupRequestFromResourceData(d), resourceNetboxContactGroupRead)
}

func resourceNetboxContactGroupRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, contactGroupAPIPath, func(res *contactGroupResponse) {
		d.Set(customFieldsKey, getCustomFields(res.CustomFields))
		d.Set("comments", res.Comments)
		d.Set("owner_id", netboxapi.RefID(res.Owner))
		d.Set("name", res.Name)
		d.Set("slug", res.Slug)
		d.Set("description", res.Description)
		if res.Parent != nil {
			d.Set("parent_id", res.Parent.ID)
		}
	})
}

func resourceNetboxContactGroupUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, contactGroupAPIPath, contactGroupRequestFromResourceData(d), resourceNetboxContactGroupRead)
}

func resourceNetboxContactGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, contactGroupAPIPath)
}
