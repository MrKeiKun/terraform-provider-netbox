package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxOwnerGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxOwnerGroupCreate),
		ReadContext:   withContext(resourceNetboxOwnerGroupRead),
		UpdateContext: withContext(resourceNetboxOwnerGroupUpdate),
		DeleteContext: withContext(resourceNetboxOwnerGroupDelete),

		Description: `:meta:subcategory:Authentication:From the [official documentation](https://docs.netbox.dev/en/stable/models/users/ownergroup/):

> An owner group is a set of owners.

Requires NetBox v4.5.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the owner group.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the owner group.",
				Optional:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const ownerGroupAPIPath = "users/owner-groups/"

type ownerGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ownerGroupResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ownerGroupRequestFromResourceData(api *providerState, d *schema.ResourceData) (*ownerGroupRequest, error) {
	data := &ownerGroupRequest{
		Name:        d.Get("name").(string),
		Description: getOptionalStr(d, "description", true),
	}
	return data, nil
}

func resourceNetboxOwnerGroupCreate(d *schema.ResourceData, m interface{}) error {
	data, err := ownerGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, ownerGroupAPIPath, data, resourceNetboxOwnerGroupRead)
}

func resourceNetboxOwnerGroupRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, ownerGroupAPIPath, func(r *ownerGroupResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
	})
}

func resourceNetboxOwnerGroupUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := ownerGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, ownerGroupAPIPath, data, resourceNetboxOwnerGroupRead)
}

func resourceNetboxOwnerGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, ownerGroupAPIPath)
}
