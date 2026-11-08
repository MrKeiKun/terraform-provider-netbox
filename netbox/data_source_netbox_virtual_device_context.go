package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxVirtualDeviceContext() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxVirtualDeviceContextRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"device_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"identifier": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"primary_ip4_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"primary_ip6_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Computed: true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxVirtualDeviceContextRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addIntQuery(query, d, "device_id", "device_id")
	r, err := findOne[map[string]interface{}](m, "dcim/virtual-device-contexts/", query, "virtual device context")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("device_id", jID(*r, "device"))
	d.Set("identifier", jIntP(*r, "identifier"))
	d.Set("tenant_id", jID(*r, "tenant"))
	d.Set("primary_ip4_id", jID(*r, "primary_ip4"))
	d.Set("primary_ip6_id", jID(*r, "primary_ip6"))
	d.Set("status", jChoice(*r, "status"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
