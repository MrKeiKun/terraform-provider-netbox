package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxRackRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxRackRoleRead),
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves a single rack role from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the rack role.",
				Required:    true,
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "The URL-friendly slug of the rack role.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the rack role.",
				Computed:    true,
			},
			"color_hex": {
				Type:        schema.TypeString,
				Description: "The color hex of the rack role.",
				Computed:    true,
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
