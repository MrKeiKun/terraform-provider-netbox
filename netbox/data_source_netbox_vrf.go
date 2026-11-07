package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxVrf() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxVrfRead,
		Description: `:meta:subcategory:IP Address Management (IPAM):`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

type vrfDataSourceResponse struct {
	ID     int64          `json:"id"`
	Name   string         `json:"name"`
	Tenant *netboxapi.Ref `json:"tenant"`
}

func dataSourceNetboxVrfRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addIntQuery(query, d, "tenant_id", "tenant_id")
	r, err := findOne[vrfDataSourceResponse](m, "ipam/vrfs/", query, "vrf")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("name", r.Name)
	d.Set("tenant_id", netboxapi.RefID(r.Tenant))
	return nil
}
