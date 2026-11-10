package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxIPAMRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxIPAMRoleRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves a single ipam role from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the ipam role.",
				Required:    true,
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "The URL-friendly slug of the ipam role.",
				Computed:    true,
			},
			"weight": {
				Type:        schema.TypeInt,
				Description: "The weight of the ipam role.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ipam role.",
				Computed:    true,
			},
		},
	}
}

type ipamRoleDataSourceResponse struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Weight      *int64 `json:"weight"`
	Description string `json:"description"`
}

func dataSourceNetboxIPAMRoleRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[ipamRoleDataSourceResponse](m, "ipam/roles/", query, "ipam role")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("slug", r.Slug)
	d.Set("name", r.Name)
	if r.Weight != nil {
		d.Set("weight", r.Weight)
	}
	if r.Description != "" {
		d.Set("description", r.Description)
	}
	return nil
}
