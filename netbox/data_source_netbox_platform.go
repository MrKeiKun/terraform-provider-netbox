package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxPlatform() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxPlatformRead,
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
			"manufacturer_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

type platformDataSourceResponse struct {
	ID           int64              `json:"id"`
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description"`
	Manufacturer *netboxapi.Ref     `json:"manufacturer"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

func dataSourceNetboxPlatformRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[platformDataSourceResponse](m, "dcim/platforms/", query, "platform")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	d.Set("slug", r.Slug)
	d.Set("description", r.Description)
	if r.Manufacturer != nil {
		d.Set("manufacturer_id", r.Manufacturer.ID)
	}
	d.Set(tagsKey, tagNames(r.Tags))
	return nil
}
