package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxContactRole() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxContactRoleRead,
		Description: `:meta:subcategory:Tenancy:`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
		},
	}
}

type contactRoleDataSourceResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func dataSourceNetboxContactRoleRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[contactRoleDataSourceResponse](m, "tenancy/contact-roles/", query, "contact role")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	d.Set("slug", r.Slug)
	return nil
}
