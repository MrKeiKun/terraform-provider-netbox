package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxExportTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxExportTemplateRead),
		Description: `:meta:subcategory:Extras:Retrieves a single export template from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"object_types": {
				Type:        schema.TypeSet,
				Description: "The object types the export template applies to.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the export template.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the export template.",
				Computed:    true,
			},
			"environment_params": {
				Type:        schema.TypeString,
				Description: "The environment params of the export template.",
				Computed:    true,
			},
			"template_code": {
				Type:        schema.TypeString,
				Description: "The template code of the export template.",
				Computed:    true,
			},
			"mime_type": {
				Type:        schema.TypeString,
				Description: "The mime type of the export template.",
				Computed:    true,
			},
			"file_name": {
				Type:        schema.TypeString,
				Description: "The file name of the export template.",
				Computed:    true,
			},
			"file_extension": {
				Type:        schema.TypeString,
				Description: "The file extension of the export template.",
				Computed:    true,
			},
			"as_attachment": {
				Type:        schema.TypeBool,
				Description: "Whether as attachment is enabled for the export template.",
				Computed:    true,
			},
			"data_source_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the data source associated with the export template.",
				Computed:    true,
			},
			"auto_sync_enabled": {
				Type:        schema.TypeBool,
				Description: "Whether auto sync enabled is enabled for the export template.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the export template.",
				Computed:    true,
			},
		},
	}
}

func dataSourceNetboxExportTemplateRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "extras/export-templates/", query, "export template")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("object_types", jStrs(*r, "object_types"))
	d.Set("name", jStr(*r, "name"))
	d.Set("description", jStr(*r, "description"))
	d.Set("environment_params", jJSON(*r, "environment_params"))
	d.Set("template_code", jStr(*r, "template_code"))
	d.Set("mime_type", jStr(*r, "mime_type"))
	d.Set("file_name", jStr(*r, "file_name"))
	d.Set("file_extension", jStr(*r, "file_extension"))
	d.Set("as_attachment", jBool(*r, "as_attachment"))
	d.Set("data_source_id", jID(*r, "data_source"))
	d.Set("auto_sync_enabled", jBool(*r, "auto_sync_enabled"))
	d.Set("owner_id", jID(*r, "owner"))
	return nil
}
