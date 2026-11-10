package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxLocation() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxLocationRead),
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves a single location from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the location.",
				Optional:    true,
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "The URL-friendly slug of the location.",
				Optional:    true,
			},
			"id": {
				Type:        schema.TypeString,
				Description: "The numeric ID of the location.",
				Optional:    true,
				Computed:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: "A description of the location.",
				Computed:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the location.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the location.",
				Computed:    true,
			},
			"site_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the site associated with the location.",
				Optional:    true,
				Computed:    true,
			},
			"parent_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the parent associated with the location.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

type locationDataSourceResponse struct {
	ID          int64             `json:"id"`
	Description string            `json:"description"`
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Site        *netboxapi.Ref    `json:"site"`
	Parent      *netboxapi.Ref    `json:"parent"`
	Status      *netboxapi.Choice `json:"status"`
	Tenant      *netboxapi.Ref    `json:"tenant"`
}

func dataSourceNetboxLocationRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "slug", "slug")
	addStrQuery(query, d, "id", "id")
	addIntQuery(query, d, "site_id", "site_id")
	addIntQuery(query, d, "parent_id", "parent_id")

	location, err := findOne[locationDataSourceResponse](m, "dcim/locations/", query, "location")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(location.ID, 10))
	d.Set("description", location.Description)
	d.Set("name", location.Name)
	if location.Site != nil {
		d.Set("site_id", location.Site.ID)
	}
	d.Set("slug", location.Slug)
	if location.Parent != nil {
		d.Set("parent_id", location.Parent.ID)
	}
	if location.Status != nil {
		d.Set("status", location.Status.Value)
	}
	if location.Tenant != nil {
		d.Set("tenant_id", location.Tenant.ID)
	}
	return nil
}
