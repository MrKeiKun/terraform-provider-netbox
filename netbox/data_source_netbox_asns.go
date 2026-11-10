package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxAsns() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxAsnsRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves all asns from NetBox. Narrow the result with filter blocks and cap it with limit.`,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeSet,
				Description: "Filters that select which items are returned. Each filter is a name/value pair.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the field to filter by.",
							Required:    true,
						},
						"value": {
							Type:        schema.TypeString,
							Description: "The value of the field to filter by.",
							Required:    true,
						},
					},
				},
			},
			"limit": {
				Type:             schema.TypeInt,
				Description:      "The maximum number of items to return. Defaults to 0, which returns all items.",
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
			},
			"asns": {
				Type:        schema.TypeList,
				Description: "The asns of the asns.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the asns.",
							Computed:    true,
						},
						"asn": {
							Type:        schema.TypeInt,
							Description: "The AS number.",
							Computed:    true,
						},
						"rir_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the rir associated with the asns.",
							Computed:    true,
						},
						"tags": tagsSchemaRead,
					},
				},
			},
		},
	}
}

func dataSourceNetboxAsnsRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"asn":      "asn",
		"asn__gte": "asn__gte",
		"asn__lte": "asn__lte",
		"asn__n":   "asn__n",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "ipam/asns/", "ASNs", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["asn"] = jIntP(v, "asn")
		mapping["rir_id"] = jID(v, "rir")
		mapping["tags"] = jTags(v)

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("asns", s)
}
