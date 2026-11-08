package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxIKEProposal() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxIKEProposalRead,
		Description: `:meta:subcategory:VPN Tunnels:Retrieves a single ike proposal from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the ike proposal.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ike proposal.",
				Computed:    true,
			},
			"authentication_method": {
				Type:        schema.TypeString,
				Description: "The authentication method of the ike proposal.",
				Computed:    true,
			},
			"encryption_algorithm": {
				Type:        schema.TypeString,
				Description: "The encryption algorithm of the ike proposal.",
				Computed:    true,
			},
			"authentication_algorithm": {
				Type:        schema.TypeString,
				Description: "The authentication algorithm of the ike proposal.",
				Computed:    true,
			},
			"group": {
				Type:        schema.TypeInt,
				Description: "The group of the ike proposal.",
				Computed:    true,
			},
			"sa_lifetime": {
				Type:        schema.TypeInt,
				Description: "The sa lifetime of the ike proposal.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the ike proposal.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the ike proposal.",
				Computed:    true,
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
