package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxVirtualCircuit() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxVirtualCircuitRead,
		Description: `:meta:subcategory:Circuits:Retrieves a single virtual circuit from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"cid": {
				Type:         schema.TypeString,
				Description:  "The circuit ID of the virtual circuit.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"cid"},
			},
			"provider_network_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the provider network associated with the virtual circuit.",
				Computed:    true,
			},
			"provider_account_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the provider account associated with the virtual circuit.",
				Computed:    true,
			},
			"type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the type associated with the virtual circuit.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the virtual circuit.",
				Computed:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the virtual circuit.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the virtual circuit.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the virtual circuit.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the virtual circuit.",
				Computed:    true,
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
