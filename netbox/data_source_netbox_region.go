package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxRegion() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxRegionRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):`,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"slug": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_region_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

type regionDataSourceResponse struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Parent      *netboxapi.Ref `json:"parent"`
}

func dataSourceNetboxRegionRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	if filter, ok := d.GetOk("filter"); ok {
		for _, f := range filter.(*schema.Set).List() {
			filterMap := f.(map[string]interface{})
			if id, ok := filterMap["id"].(int); ok && id != 0 {
				query.Set("id", strconv.Itoa(id))
			}
			if name, ok := filterMap["name"].(string); ok && name != "" {
				query.Set("name", name)
			}
			if slug, ok := filterMap["slug"].(string); ok && slug != "" {
				query.Set("slug", slug)
			}
		}
	}

	result, err := findOne[regionDataSourceResponse](m, "dcim/regions/", query, "region")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(result.ID, 10))
	d.Set("name", result.Name)
	d.Set("slug", result.Slug)
	d.Set("description", result.Description)
	if result.Parent != nil {
		d.Set("parent_region_id", result.Parent.ID)
	}
	return nil
}
