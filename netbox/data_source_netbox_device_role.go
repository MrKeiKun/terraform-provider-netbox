package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxDeviceRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxDeviceRoleRead),
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves a single device role from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the device role.",
				Required:    true,
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "The URL-friendly slug of the device role.",
				Computed:    true,
			},
			"color_hex": {
				Type:        schema.TypeString,
				Description: "The color hex of the device role.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

type deviceRoleDataSourceResponse struct {
	ID    int64              `json:"id"`
	Name  string             `json:"name"`
	Slug  string             `json:"slug"`
	Color string             `json:"color"`
	Tags  []netboxapi.TagRef `json:"tags"`
}

func dataSourceNetboxDeviceRoleRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[deviceRoleDataSourceResponse](m, "dcim/device-roles/", query, "device role")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	d.Set("slug", r.Slug)
	d.Set("color_hex", r.Color)
	d.Set(tagsKey, tagNames(r.Tags))
	return nil
}
