package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxIKEProposal() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxIKEProposalRead,
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
			"authentication_method": {
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
			"group": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"sa_lifetime": {
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

func dataSourceNetboxIKEProposalRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "vpn/ike-proposals/", query, "ike proposal")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("description", jStr(*r, "description"))
	d.Set("authentication_method", jChoice(*r, "authentication_method"))
	d.Set("encryption_algorithm", jChoice(*r, "encryption_algorithm"))
	d.Set("authentication_algorithm", jChoice(*r, "authentication_algorithm"))
	d.Set("group", jChoiceInt(*r, "group"))
	d.Set("sa_lifetime", jIntP(*r, "sa_lifetime"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
