package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxRackRole() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxRackRoleRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):`,
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
				Computed: true,
			},
			"color_hex": {
				Type:     schema.TypeString,
				Computed: true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

type rackRoleDataSourceResponse struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description"`
	Color       string             `json:"color"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func dataSourceNetboxRackRoleRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[rackRoleDataSourceResponse](m, "dcim/rack-roles/", query, "rack role")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	d.Set("slug", r.Slug)
	d.Set("description", r.Description)
	d.Set("color_hex", r.Color)
	d.Set(tagsKey, tagNames(r.Tags))
	return nil
}
