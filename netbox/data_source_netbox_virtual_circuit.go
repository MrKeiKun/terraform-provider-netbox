package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxVirtualCircuit() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxVirtualCircuitRead,
		Description: `:meta:subcategory:Circuits:`,
		Schema: map[string]*schema.Schema{
			"cid": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"cid"},
			},
			"provider_network_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"provider_account_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"type_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
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

func dataSourceNetboxVirtualCircuitRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "cid", "cid")
	r, err := findOne[map[string]interface{}](m, "circuits/virtual-circuits/", query, "virtual circuit")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("cid", jStr(*r, "cid"))
	d.Set("provider_network_id", jID(*r, "provider_network"))
	d.Set("provider_account_id", jID(*r, "provider_account"))
	d.Set("type_id", jID(*r, "type"))
	d.Set("status", jChoice(*r, "status"))
	d.Set("tenant_id", jID(*r, "tenant"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
