package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxPlatform() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxPlatformRead),
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves a single platform from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the platform.",
				Required:    true,
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "The URL-friendly slug of the platform.",
				Computed:    true,
			},
			"manufacturer_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the manufacturer associated with the platform.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the platform.",
				Computed:    true,
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
