package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxTenantGroup() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxTenantGroupRead,
		Description: `:meta:subcategory:Tenancy:Retrieves a single tenant group from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the tenant group.",
				Required:    true,
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "The URL-friendly slug of the tenant group.",
				Computed:    true,
			},
			"parent_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the parent associated with the tenant group.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the tenant group.",
				Computed:    true,
			},
		},
	}
}

type tenantGroupDataSourceResponse struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Parent      *netboxapi.Ref `json:"parent"`
}

func dataSourceNetboxTenantGroupRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[tenantGroupDataSourceResponse](m, "tenancy/tenant-groups/", query, "tenant group")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	d.Set("slug", r.Slug)
	d.Set("description", r.Description)
	if r.Parent != nil {
		d.Set("parent_id", r.Parent.ID)
	}
	return nil
}
