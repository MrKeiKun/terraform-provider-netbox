package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxOwner() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxOwnerRead,
		Description: `:meta:subcategory:Authentication:Requires NetBox v4.5.0 or newer.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the owner.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the group associated with the owner.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the owner.",
				Computed:    true,
			},
			"user_group_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the user group associated with the owner.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"user_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the user associated with the owner.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func dataSourceNetboxOwnerRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "users/owners/", query, "owner")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("group_id", jID(*r, "group"))
	d.Set("description", jStr(*r, "description"))
	d.Set("user_group_ids", jIDs(*r, "user_groups"))
	d.Set("user_ids", jIDs(*r, "users"))
	return nil
}
