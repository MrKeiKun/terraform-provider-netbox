package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxTag() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxTagRead,
		Description: `:meta:subcategory:Extras:`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

type tagDataSourceResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func dataSourceNetboxTagRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[tagDataSourceResponse](m, "extras/tags/", query, "tag")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	d.Set("slug", r.Slug)
	d.Set("description", r.Description)
	return nil
}
