package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxServiceTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxServiceTemplateRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves a single service template from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the service template.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"port_mappings": {
				Type:        schema.TypeSet,
				Description: "The port mappings of the service template.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"protocol": {
				Type:        schema.TypeString,
				Description: "The protocol of the service template.",
				Computed:    true,
			},
			"ports": {
				Type:        schema.TypeSet,
				Description: "The ports of the service template.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the service template.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the service template.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the service template.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxServiceTemplateRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "ipam/service-templates/", query, "service template")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("port_mappings", jStrs(*r, "port_mappings"))
	d.Set("protocol", jChoice(*r, "protocol"))
	d.Set("ports", jInts(*r, "ports"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
