package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxGroupCreate),
		ReadContext:   withContext(resourceNetboxGroupRead),
		UpdateContext: withContext(resourceNetboxGroupUpdate),
		DeleteContext: withContext(resourceNetboxGroupDelete),

		Description: `:meta:subcategory:Authentication:This resource is used to manage groups.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the group.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the group.",
				Optional:    true,
				Default:     "",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const groupAPIPath = "users/groups/"

type groupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type groupResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func groupRequestFromResourceData(d *schema.ResourceData) *groupRequest {
	name := d.Get("name").(string)
	data := &groupRequest{
		Name:        name,
		Description: d.Get("description").(string),
	}
	return data
}

func resourceNetboxGroupCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, groupAPIPath, groupRequestFromResourceData(d), resourceNetboxGroupRead)
}

func resourceNetboxGroupRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, groupAPIPath, func(r *groupResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
	})
}

func resourceNetboxGroupUpdate(d *schema.ResourceData, m interface{}) error {
	return putObject(d, m, groupAPIPath, groupRequestFromResourceData(d), resourceNetboxGroupRead)
}

func resourceNetboxGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, groupAPIPath)
}
