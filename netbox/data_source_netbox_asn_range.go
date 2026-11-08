package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxASNRange() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxASNRangeRead,
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves a single asn range from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the asn range.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the asn range.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"rir_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the rir associated with the asn range.",
				Computed:    true,
			},
			"start": {
				Type:        schema.TypeInt,
				Description: "The start of the asn range.",
				Computed:    true,
			},
			"end": {
				Type:        schema.TypeInt,
				Description: "The end of the asn range.",
				Computed:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the asn range.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the asn range.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the asn range.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the asn range.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxASNRangeRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "slug", "slug")
	r, err := findOne[map[string]interface{}](m, "ipam/asn-ranges/", query, "asn range")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("slug", jStr(*r, "slug"))
	d.Set("rir_id", jID(*r, "rir"))
	d.Set("start", jIntP(*r, "start"))
	d.Set("end", jIntP(*r, "end"))
	d.Set("tenant_id", jID(*r, "tenant"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
