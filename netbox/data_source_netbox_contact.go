package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxContact() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxContactRead,
		Description: `:meta:subcategory:Tenancy:Retrieves a single contact from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the contact.",
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the contact.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the group associated with the contact.",
				Deprecated:  "This field is deprecated. Please use the new \"group_ids\" attribute instead.",
				Computed:    true,
			},
			"group_ids": {
				Type:        schema.TypeList,
				Description: "The IDs of the group associated with the contact.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Computed: true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the contact.",
				Optional:    true,
			},
		},
	}
}

type contactDataSourceResponse struct {
	ID     int64           `json:"id"`
	Name   string          `json:"name"`
	Groups []netboxapi.Ref `json:"groups"`
}

func dataSourceNetboxContactRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")

	r, err := findOne[contactDataSourceResponse](m, "tenancy/contacts/", query, "contact")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	if r.Groups != nil {
		groupIDs := netboxapi.IDs(r.Groups)
		d.Set("group_ids", groupIDs)
		if len(groupIDs) > 0 {
			d.Set("group_id", groupIDs[0])
		}
	}
	return nil
}
