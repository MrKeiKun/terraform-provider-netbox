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

const contactRoleAPIPath = "tenancy/contact-roles/"

type contactRoleRequest struct {
	Name string             `json:"name"`
	Slug string             `json:"slug"`
	Tags []netboxapi.TagRef `json:"tags"`
}

type contactRoleResponse struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func contactRoleRequestFromResourceData(d *schema.ResourceData) *contactRoleRequest {
	name := d.Get("name").(string)
	return &contactRoleRequest{
		Name: name,
		Slug: slugFromResourceData(d, name),
		Tags: []netboxapi.TagRef{},
	}
}

func resourceNetboxContactRoleCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, contactRoleAPIPath, contactRoleRequestFromResourceData(d), resourceNetboxContactRoleRead)
}

func resourceNetboxContactRoleRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, contactRoleAPIPath, func(res *contactRoleResponse) {
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
