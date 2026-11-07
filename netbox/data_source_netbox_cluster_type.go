package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxClusterType() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxClusterTypeRead,
		Description: `:meta:subcategory:Virtualization:`,
		Schema: map[string]*schema.Schema{
			"cluster_type_id": {
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

type clusterTypeDataSourceResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func dataSourceNetboxClusterTypeRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[clusterTypeDataSourceResponse](m, "virtualization/cluster-types/", query, "cluster type")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("cluster_type_id", r.ID)
	d.Set("name", r.Name)
	return nil
}
