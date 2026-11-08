package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxCustomLink() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxCustomLinkRead,
		Description: `:meta:subcategory:Extras:`,
		Schema: map[string]*schema.Schema{
			"object_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"link_text": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"link_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weight": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"button_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"new_window": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceNetboxCustomLinkRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "extras/custom-links/", query, "custom link")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("object_types", jStrs(*r, "object_types"))
	d.Set("name", jStr(*r, "name"))
	d.Set("enabled", jBool(*r, "enabled"))
	d.Set("link_text", jStr(*r, "link_text"))
	d.Set("link_url", jStr(*r, "link_url"))
	d.Set("weight", jIntP(*r, "weight"))
	d.Set("group_name", jStr(*r, "group_name"))
	d.Set("button_class", jStr(*r, "button_class"))
	d.Set("new_window", jBool(*r, "new_window"))
	d.Set("owner_id", jID(*r, "owner"))
	return nil
}
