package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxL2VPN() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxL2VPNRead),
		Description: `:meta:subcategory:VPN Tunnels:Retrieves a single l2vpn from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"identifier": {
				Type:        schema.TypeInt,
				Description: "The identifier of the l2vpn.",
				Computed:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the l2vpn.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the l2vpn.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the l2vpn.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the l2vpn.",
				Computed:    true,
			},
			"import_target_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the import target associated with the l2vpn.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"export_target_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the export target associated with the l2vpn.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the l2vpn.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the l2vpn.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the l2vpn.",
				Computed:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the l2vpn.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxL2VPNRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "slug", "slug")
	r, err := findOne[map[string]interface{}](m, "vpn/l2vpns/", query, "l2vpn")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("identifier", jIntP(*r, "identifier"))
	d.Set("name", jStr(*r, "name"))
	d.Set("slug", jStr(*r, "slug"))
	d.Set("type", jChoice(*r, "type"))
	d.Set("status", jChoice(*r, "status"))
	d.Set("import_target_ids", jIDs(*r, "import_targets"))
	d.Set("export_target_ids", jIDs(*r, "export_targets"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set("tenant_id", jID(*r, "tenant"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
