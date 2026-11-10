package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxTag() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxTagRead),
		Description: `:meta:subcategory:Extras:Retrieves a single tag from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the tag.",
				Required:    true,
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "The URL-friendly slug of the tag.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the tag.",
				Optional:    true,
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
