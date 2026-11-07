package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxClusterGroup() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxClusterGroupRead,
		Description: `:meta:subcategory:Virtualization:`,
		Schema: map[string]*schema.Schema{
			"cluster_group_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

type clusterGroupDataSourceResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func dataSourceNetboxClusterGroupRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[clusterGroupDataSourceResponse](m, "virtualization/cluster-groups/", query, "cluster group")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("cluster_group_id", r.ID)
	d.Set("name", r.Name)
	return nil
}
