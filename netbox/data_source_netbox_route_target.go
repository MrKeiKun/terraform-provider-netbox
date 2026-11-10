package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxRouteTarget() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxRouteTargetRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves a single route target from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the route target.",
				ValidateFunc: validation.StringLenBetween(1, 21),
				Required:     true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the route target.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the route target.",
				Computed:    true,
			},
			tagsKey: tagsSchema,
		},
	}
}

type routeTargetDataSourceResponse struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Tenant      *netboxapi.Ref     `json:"tenant"`
	Description string             `json:"description"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func dataSourceNetboxRouteTargetRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[routeTargetDataSourceResponse](m, "ipam/route-targets/", query, "route target")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	if r.Tenant != nil {
		d.Set("tenant_id", r.Tenant.ID)
	}
	if r.Description != "" {
		d.Set("description", r.Description)
	}
	d.Set(tagsKey, tagNames(r.Tags))
	return nil
}
