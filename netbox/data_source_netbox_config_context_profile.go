package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxConfigContextProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxConfigContextProfileRead),
		Description: `:meta:subcategory:Extras:Retrieves a single config context profile from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the config context profile.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the config context profile.",
				Computed:    true,
			},
			"schema": {
				Type:        schema.TypeString,
				Description: "The schema of the config context profile.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the config context profile.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the config context profile.",
				Computed:    true,
			},
			"data_source_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the data source associated with the config context profile.",
				Computed:    true,
			},
		},
	}
}

func dataSourceNetboxConfigContextProfileRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "extras/config-context-profiles/", query, "config context profile")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("description", jStr(*r, "description"))
	d.Set("schema", jJSON(*r, "schema"))
	d.Set(tagsKey, jTags(*r))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set("data_source_id", jID(*r, "data_source"))
	return nil
}
