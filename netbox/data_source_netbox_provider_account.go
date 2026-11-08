package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxProviderAccount() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxProviderAccountRead,
		Description: `:meta:subcategory:Circuits:Retrieves a single provider account from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"provider_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the provider associated with the provider account.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the provider account.",
				Computed:    true,
			},
			"account": {
				Type:         schema.TypeString,
				Description:  "The account of the provider account.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"account"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the provider account.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the provider account.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the provider account.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxProviderAccountRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "account", "account")
	r, err := findOne[map[string]interface{}](m, "circuits/provider-accounts/", query, "provider account")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("provider_id", jID(*r, "provider"))
	d.Set("name", jStr(*r, "name"))
	d.Set("account", jStr(*r, "account"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
