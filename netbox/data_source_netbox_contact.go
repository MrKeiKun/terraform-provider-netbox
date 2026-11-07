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
		Description: `:meta:subcategory:Tenancy:`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"group_id": {
				Type:       schema.TypeInt,
				Deprecated: "This field is deprecated. Please use the new \"group_ids\" attribute instead.",
				Computed:   true,
			},
			"group_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
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
