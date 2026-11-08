package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxIPSecProfile() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxIPSecProfileRead,
		Description: `:meta:subcategory:VPN Tunnels:`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ike_policy_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ipsec_policy_id": {
				Type:     schema.TypeInt,
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

func dataSourceNetboxIPSecProfileRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "vpn/ipsec-profiles/", query, "ipsec profile")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("description", jStr(*r, "description"))
	d.Set("mode", jChoice(*r, "mode"))
	d.Set("ike_policy_id", jID(*r, "ike_policy"))
	d.Set("ipsec_policy_id", jID(*r, "ipsec_policy"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
