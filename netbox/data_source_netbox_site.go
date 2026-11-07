package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxSite() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxSiteRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"facility": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"asn_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"comments": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"site_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"time_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"physical_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

type siteDataSourceResponse struct {
	ID              int64             `json:"id"`
	Asns            []netboxapi.Ref   `json:"asns"`
	Comments        string            `json:"comments"`
	Description     string            `json:"description"`
	Name            string            `json:"name"`
	Slug            string            `json:"slug"`
	TimeZone        *string           `json:"time_zone"`
	Facility        string            `json:"facility"`
	PhysicalAddress string            `json:"physical_address"`
	Group           *netboxapi.Ref    `json:"group"`
	Region          *netboxapi.Ref    `json:"region"`
	Status          *netboxapi.Choice `json:"status"`
	Tenant          *netboxapi.Ref    `json:"tenant"`
}

func dataSourceNetboxSiteRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "slug", "slug")
	if id, ok := d.Get("id").(string); ok && id != "0" {
		addStrQuery(query, d, "id", "id")
	}
	addStrQuery(query, d, "facility", "facility")

	site, err := findOne[siteDataSourceResponse](m, "dcim/sites/", query, "site")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(site.ID, 10))
	d.Set("asn_ids", netboxapi.IDs(site.Asns))
	d.Set("comments", site.Comments)
	d.Set("description", site.Description)
	d.Set("name", site.Name)
	d.Set("site_id", site.ID)
	d.Set("slug", site.Slug)
	d.Set("time_zone", site.TimeZone)
	d.Set("facility", site.Facility)
	d.Set("physical_address", site.PhysicalAddress)

	if site.Group != nil {
		d.Set("group_id", site.Group.ID)
	}
	if site.Region != nil {
		d.Set("region_id", site.Region.ID)
	}
	if site.Status != nil {
		d.Set("status", site.Status.Value)
	}
	if site.Tenant != nil {
		d.Set("tenant_id", site.Tenant.ID)
	}

	return nil
}
