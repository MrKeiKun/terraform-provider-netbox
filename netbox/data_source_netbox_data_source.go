package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxDataSourceRead),
		Description: `:meta:subcategory:Extras:Retrieves a single data source from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the data source.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the data source.",
				Computed:    true,
			},
			"source_url": {
				Type:        schema.TypeString,
				Description: "The source url of the data source.",
				Computed:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the data source is enabled.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the data source.",
				Computed:    true,
			},
			"sync_interval": {
				Type:        schema.TypeInt,
				Description: "The sync interval of the data source.",
				Computed:    true,
			},
			"parameters": {
				Type:        schema.TypeString,
				Description: "The parameters of the data source.",
				Computed:    true,
			},
			"ignore_rules": {
				Type:        schema.TypeString,
				Description: "The ignore rules of the data source.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the data source.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the data source.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxDataSourceRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "core/data-sources/", query, "data source")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("type", jChoice(*r, "type"))
	d.Set("source_url", jStr(*r, "source_url"))
	d.Set("enabled", jBool(*r, "enabled"))
	d.Set("description", jStr(*r, "description"))
	d.Set("sync_interval", jIntP(*r, "sync_interval"))
	d.Set("parameters", jJSON(*r, "parameters"))
	d.Set("ignore_rules", jStr(*r, "ignore_rules"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
