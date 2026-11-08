package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxTenant() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxTenantRead,
		Description: `:meta:subcategory:Tenancy:Retrieves a single tenant from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the tenant.",
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the tenant.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the group associated with the tenant.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the tenant.",
				Optional:    true,
			},
			"tags": tagsSchemaRead,
		},
	}
}

type tenantDataSourceResponse struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description"`
	Group       *netboxapi.Ref     `json:"group"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func dataSourceNetboxTenantRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "slug", "slug")
	r, err := findOne[tenantDataSourceResponse](m, "tenancy/tenants/", query, "tenant")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	d.Set("slug", r.Slug)
	d.Set("description", r.Description)
	if r.Group != nil {
		d.Set("group_id", r.Group.ID)
	}
	d.Set("tags", tagNames(r.Tags))
	return nil
}
