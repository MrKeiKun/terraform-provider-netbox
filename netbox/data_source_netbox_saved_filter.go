package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxSavedFilter() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxSavedFilterRead),
		Description: `:meta:subcategory:Extras:Retrieves a single saved filter from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"object_types": {
				Type:        schema.TypeSet,
				Description: "The object types the saved filter applies to.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the saved filter.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the saved filter.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the saved filter.",
				Computed:    true,
			},
			"user_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the user associated with the saved filter.",
				Computed:    true,
			},
			"weight": {
				Type:        schema.TypeInt,
				Description: "The weight of the saved filter.",
				Computed:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the saved filter is enabled.",
				Computed:    true,
			},
			"shared": {
				Type:        schema.TypeBool,
				Description: "Whether shared is enabled for the saved filter.",
				Computed:    true,
			},
			"parameters": {
				Type:        schema.TypeString,
				Description: "The parameters of the saved filter.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the saved filter.",
				Computed:    true,
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
