package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxCoolingFeed() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxCoolingFeedRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Requires NetBox v4.7.0 or newer.`,
		Schema: map[string]*schema.Schema{
			"cooling_source_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the cooling source associated with the cooling feed.",
				Computed:    true,
			},
			"rack_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the rack associated with the cooling feed.",
				Computed:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the cooling feed.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the cooling feed.",
				Computed:    true,
			},
			"cooling_capacity": {
				Type:        schema.TypeFloat,
				Description: "The cooling capacity of the cooling feed.",
				Computed:    true,
			},
			"max_flow": {
				Type:        schema.TypeFloat,
				Description: "The max flow of the cooling feed.",
				Computed:    true,
			},
			"max_flow_unit": {
				Type:        schema.TypeString,
				Description: "The max flow unit of the cooling feed.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cooling feed.",
				Computed:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the cooling feed.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the cooling feed.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the cooling feed.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxCoolingFeedRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "dcim/cooling-feeds/", query, "cooling feed")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("cooling_source_id", jID(*r, "cooling_source"))
	d.Set("rack_id", jID(*r, "rack"))
	d.Set("name", jStr(*r, "name"))
	d.Set("status", jChoice(*r, "status"))
	d.Set("cooling_capacity", jFloatP(*r, "cooling_capacity"))
	d.Set("max_flow", jFloatP(*r, "max_flow"))
	d.Set("max_flow_unit", jChoice(*r, "max_flow_unit"))
	d.Set("description", jStr(*r, "description"))
	d.Set("tenant_id", jID(*r, "tenant"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
