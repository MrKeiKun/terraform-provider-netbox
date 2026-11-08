package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxSavedFilter() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxSavedFilterRead,
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
				AtLeastOneOf: []string{"name", "slug"},
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"weight": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"parameters": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceNetboxSavedFilterRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "slug", "slug")
	r, err := findOne[map[string]interface{}](m, "extras/saved-filters/", query, "saved filter")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("object_types", jStrs(*r, "object_types"))
	d.Set("name", jStr(*r, "name"))
	d.Set("slug", jStr(*r, "slug"))
	d.Set("description", jStr(*r, "description"))
	d.Set("user_id", jID(*r, "user"))
	d.Set("weight", jIntP(*r, "weight"))
	d.Set("enabled", jBool(*r, "enabled"))
	d.Set("shared", jBool(*r, "shared"))
	d.Set("parameters", jJSON(*r, "parameters"))
	d.Set("owner_id", jID(*r, "owner"))
	return nil
}
