package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxAsn() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxAsnRead,
		Description: `:meta:subcategory:IP Address Management (IPAM):`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"asn": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"asn", "tag"},
			},
			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"asn", "tag"},
				Description:  "Tag to include in the data source filter (must match the tag's slug).",
			},
			"tag__n": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `Tag to exclude from the data source filter (must match the tag's slug).
Refer to [Netbox's documentation](https://demo.netbox.dev/static/docs/rest-api/filtering/#lookup-expressions)
for more information on available lookup expressions.`,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaRead,
		},
	}
}

type asnDataSourceResponse struct {
	ID          int64              `json:"id"`
	Asn         *int64             `json:"asn"`
	Description string             `json:"description"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func dataSourceNetboxAsnRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "asn", "asn")
	addStrQuery(query, d, "tag", "tag")
	addStrQuery(query, d, "tag__n", "tag__n")

	r, err := findOne[asnDataSourceResponse](m, "ipam/asns/", query, "asn")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("id", r.ID)
	if r.Asn != nil {
		d.Set("asn", strconv.FormatInt(*r.Asn, 10))
	}
	d.Set("description", r.Description)
	d.Set("tags", tagNames(r.Tags))
	return nil
}
