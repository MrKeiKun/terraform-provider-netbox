package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxIPSecProposal() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxIPSecProposalRead,
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
			"encryption_algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sa_lifetime_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"sa_lifetime_data": {
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

func dataSourceNetboxIPSecProposalRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "vpn/ipsec-proposals/", query, "ipsec proposal")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("description", jStr(*r, "description"))
	d.Set("encryption_algorithm", jChoice(*r, "encryption_algorithm"))
	d.Set("authentication_algorithm", jChoice(*r, "authentication_algorithm"))
	d.Set("sa_lifetime_seconds", jIntP(*r, "sa_lifetime_seconds"))
	d.Set("sa_lifetime_data", jIntP(*r, "sa_lifetime_data"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
