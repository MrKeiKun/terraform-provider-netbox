package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxContactGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxContactGroupCreate,
		Read:   resourceNetboxContactGroupRead,
		Update: resourceNetboxContactGroupUpdate,
		Delete: resourceNetboxContactGroupDelete,

		Description: `:meta:subcategory:Tenancy:From the [official documentation](https://docs.netbox.dev/en/stable/features/contacts/#contact-groups):

> Contacts can be grouped arbitrarily into a recursive hierarchy, and a contact can be assigned to a group at any level within the hierarchy.`,

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
			"parent_id": {
				Type:     schema.TypeInt,
				Optional: true,
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

const contactGroupAPIPath = "tenancy/contact-groups/"

type contactGroupRequest struct {
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description,omitempty"`
	Parent      *int64             `json:"parent,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type contactGroupResponse struct {
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Parent      *netboxapi.Ref `json:"parent"`
}

func contactGroupRequestFromResourceData(d *schema.ResourceData) *contactGroupRequest {
	name := d.Get("name").(string)
	data := &contactGroupRequest{
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Description: d.Get("description").(string),
		Tags:        []netboxapi.TagRef{},
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
