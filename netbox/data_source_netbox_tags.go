package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxTags() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxTagsRead,
		Description: `:meta:subcategory:Extras:Retrieves all tags from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
			"tags": {
				Type:        schema.TypeList,
				Description: "Tags to assign to the tags.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tag_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the tag associated with the tags.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the tags.",
							Computed:    true,
						},
						"slug": {
							Type:        schema.TypeString,
							Description: "The URL-friendly slug of the tags.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the tags.",
							Optional:    true,
						},
						"color": {
							Type:        schema.TypeString,
							Description: "The color of the tags as a hex string without the leading #.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxTagsRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"id":         "id",
		"id__gt":     "id__gt",
		"id__gte":    "id__gte",
		"id__lt":     "id__lt",
		"id__lte":    "id__lte",
		"name":       "name",
		"name__ic":   "name__ic",
		"name__niew": "name__niew",
		"name__iew":  "name__iew",
		"name__nisw": "name__nisw",
		"name__isw":  "name__isw",
		"slug":       "slug",
		"slug__ic":   "slug__ic",
		"slug__niew": "slug__niew",
		"slug__iew":  "slug__iew",
		"slug__nisw": "slug__nisw",
		"slug__isw":  "slug__isw",
	}

	results, err := listFiltered[map[string]interface{}](d, m, "extras/tags/", "tags", allowed, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["tag_id"] = jInt(v, "id")
		mapping["name"] = jStr(v, "name")
		mapping["slug"] = jStr(v, "slug")
		mapping["description"] = jStr(v, "description")
		mapping["color"] = jStr(v, "color")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("tags", s)
}
