package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxIKEPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxIKEPolicyRead),
		Description: `:meta:subcategory:VPN Tunnels:Retrieves a single ike policy from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the ike policy.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ike policy.",
				Computed:    true,
			},
			"version": {
				Type:        schema.TypeInt,
				Description: "The version of the ike policy.",
				Computed:    true,
			},
			"mode": {
				Type:        schema.TypeString,
				Description: "The mode of the ike policy.",
				Computed:    true,
			},
			"proposal_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the proposal associated with the ike policy.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"preshared_key": {
				Type:        schema.TypeString,
				Description: "The preshared key of the ike policy.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the ike policy.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the ike policy.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxIKEPolicyRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	r, err := findOne[map[string]interface{}](m, "vpn/ike-policies/", query, "ike policy")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("description", jStr(*r, "description"))
	d.Set("version", jChoiceInt(*r, "version"))
	d.Set("mode", jChoice(*r, "mode"))
	d.Set("proposal_ids", jIDs(*r, "proposals"))
	d.Set("preshared_key", jStr(*r, "preshared_key"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
