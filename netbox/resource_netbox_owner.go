package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxOwner() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxOwnerCreate,
		Read:   resourceNetboxOwnerRead,
		Update: resourceNetboxOwnerUpdate,
		Delete: resourceNetboxOwnerDelete,

		Description: `:meta:subcategory:Authentication:From the [official documentation](https://docs.netbox.dev/en/stable/models/users/owner/):

> An owner is a set of users and groups that is responsible for objects. Objects point to their owner with the owner_id attribute.

Requires NetBox v4.5.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"user_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"user_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const ownerAPIPath = "users/owners/"

type ownerRequest struct {
	Name        string  `json:"name"`
	Group       *int64  `json:"group"`
	Description string  `json:"description,omitempty"`
	UserGroups  []int64 `json:"user_groups"`
	Users       []int64 `json:"users"`
}

type ownerResponse struct {
	Name        string          `json:"name"`
	Group       *netboxapi.Ref  `json:"group"`
	Description string          `json:"description"`
	UserGroups  []netboxapi.Ref `json:"user_groups"`
	Users       []netboxapi.Ref `json:"users"`
}

func ownerRequestFromResourceData(api *providerState, d *schema.ResourceData) (*ownerRequest, error) {
	data := &ownerRequest{
		Name:        d.Get("name").(string),
		Group:       getOptionalInt(d, "group_id"),
		Description: getOptionalStr(d, "description", true),
		UserGroups:  getIDSet(d, "user_group_ids"),
		Users:       getIDSet(d, "user_ids"),
	}
	return data, nil
}

func resourceNetboxOwnerCreate(d *schema.ResourceData, m interface{}) error {
	data, err := ownerRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, ownerAPIPath, data, resourceNetboxOwnerRead)
}

func resourceNetboxOwnerRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, ownerAPIPath, func(r *ownerResponse) {
		d.Set("name", r.Name)
		d.Set("group_id", netboxapi.RefID(r.Group))
		d.Set("description", r.Description)
		d.Set("user_group_ids", netboxapi.IDs(r.UserGroups))
		d.Set("user_ids", netboxapi.IDs(r.Users))
	})
}

func resourceNetboxOwnerUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := ownerRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, ownerAPIPath, data, resourceNetboxOwnerRead)
}

func resourceNetboxOwnerDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, ownerAPIPath)
}
