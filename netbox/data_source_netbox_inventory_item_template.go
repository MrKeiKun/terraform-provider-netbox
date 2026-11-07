package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxInventoryItemTemplate() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxInventoryItemTemplateRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):`,
		Schema: map[string]*schema.Schema{
			"device_type_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"parent_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"manufacturer_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"part_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"component_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"component_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceNetboxInventoryItemTemplateRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addIntQuery(query, d, "device_type_id", "device_type_id")
	r, err := findOne[map[string]interface{}](m, "dcim/inventory-item-templates/", query, "inventory item template")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("device_type_id", jID(*r, "device_type"))
	d.Set("parent_id", jID(*r, "parent"))
	d.Set("name", jStr(*r, "name"))
	d.Set("label", jStr(*r, "label"))
	d.Set("role_id", jID(*r, "role"))
	d.Set("manufacturer_id", jID(*r, "manufacturer"))
	d.Set("part_id", jStr(*r, "part_id"))
	d.Set("description", jStr(*r, "description"))
	d.Set("component_type", jStr(*r, "component_type"))
	d.Set("component_id", jIntP(*r, "component_id"))
	return nil
}
