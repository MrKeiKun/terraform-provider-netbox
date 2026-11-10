package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxIPSecPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxIPSecPolicyRead),
		Description: `:meta:subcategory:VPN Tunnels:Retrieves a single ipsec policy from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the ipsec policy.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ipsec policy.",
				Computed:    true,
			},
			"proposal_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the proposal associated with the ipsec policy.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"pfs_group": {
				Type:        schema.TypeInt,
				Description: "The pfs group of the ipsec policy.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the ipsec policy.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the ipsec policy.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxIPSecPolicyRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "vpn/ipsec-policies/", query, "ipsec policy")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("description", jStr(*r, "description"))
	d.Set("proposal_ids", jIDs(*r, "proposals"))
	d.Set("pfs_group", jChoiceInt(*r, "pfs_group"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
