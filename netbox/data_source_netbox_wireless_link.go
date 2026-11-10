package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxWirelessLink() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxWirelessLinkRead),
		Description: `:meta:subcategory:Wireless:Retrieves a single wireless link from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"interface_a_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the interface a associated with the wireless link.",
				Computed:    true,
			},
			"interface_b_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the interface b associated with the wireless link.",
				Computed:    true,
			},
			"ssid": {
				Type:         schema.TypeString,
				Description:  "The ssid of the wireless link.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"ssid"},
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the wireless link.",
				Computed:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the wireless link.",
				Computed:    true,
			},
			"auth_type": {
				Type:        schema.TypeString,
				Description: "The auth type of the wireless link.",
				Computed:    true,
			},
			"auth_cipher": {
				Type:        schema.TypeString,
				Description: "The auth cipher of the wireless link.",
				Computed:    true,
			},
			"auth_psk": {
				Type:        schema.TypeString,
				Description: "The auth psk of the wireless link.",
				Computed:    true,
			},
			"distance": {
				Type:        schema.TypeFloat,
				Description: "The distance of the wireless link.",
				Computed:    true,
			},
			"distance_unit": {
				Type:        schema.TypeString,
				Description: "The distance unit of the wireless link.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the wireless link.",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the wireless link.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the wireless link.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxWirelessLinkRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "ssid", "ssid")
	r, err := findOne[map[string]interface{}](m, "wireless/wireless-links/", query, "wireless link")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("interface_a_id", jID(*r, "interface_a"))
	d.Set("interface_b_id", jID(*r, "interface_b"))
	d.Set("ssid", jStr(*r, "ssid"))
	d.Set("status", jChoice(*r, "status"))
	d.Set("tenant_id", jID(*r, "tenant"))
	d.Set("auth_type", jChoice(*r, "auth_type"))
	d.Set("auth_cipher", jChoice(*r, "auth_cipher"))
	d.Set("auth_psk", jStr(*r, "auth_psk"))
	d.Set("distance", jFloatP(*r, "distance"))
	d.Set("distance_unit", jChoice(*r, "distance_unit"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
