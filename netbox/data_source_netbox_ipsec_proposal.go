package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxIPSecProposal() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxIPSecProposalRead),
		Description: `:meta:subcategory:VPN Tunnels:Retrieves a single ipsec proposal from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the ipsec proposal.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ipsec proposal.",
				Computed:    true,
			},
			"encryption_algorithm": {
				Type:        schema.TypeString,
				Description: "The encryption algorithm of the ipsec proposal.",
				Computed:    true,
			},
			"authentication_algorithm": {
				Type:        schema.TypeString,
				Description: "The authentication algorithm of the ipsec proposal.",
				Computed:    true,
			},
			"sa_lifetime_seconds": {
				Type:        schema.TypeInt,
				Description: "The sa lifetime seconds of the ipsec proposal.",
				Computed:    true,
			},
			"sa_lifetime_data": {
				Type:        schema.TypeInt,
				Description: "The sa lifetime data of the ipsec proposal.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the ipsec proposal.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the ipsec proposal.",
				Computed:    true,
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
