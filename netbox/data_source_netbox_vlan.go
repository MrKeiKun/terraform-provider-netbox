package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxVlan() *schema.Resource {
	customFieldsFilterSchema := *customFieldsSchema
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxVlanRead),
		Description: `:meta:subcategory:IP Address Management (IPAM):Retrieves a single vlan from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"vid": {
				Type:         schema.TypeInt,
				Description:  "The VLAN ID (1-4094).",
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the vlan.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the vlan.",
				Computed:    true,
			},
			"group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the group associated with the vlan.",
				Computed:    true,
				Optional:    true,
			},
			"role": {
				Type:        schema.TypeInt,
				Description: "The role of the vlan.",
				Computed:    true,
				Optional:    true,
			},
			"site": {
				Type:        schema.TypeInt,
				Description: "The site of the vlan.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the vlan.",
				Computed:    true,
			},
			"tenant": {
				Type:        schema.TypeInt,
				Description: "The tenant of the vlan.",
				Computed:    true,
				Optional:    true,
			},
			customFieldsKey: &customFieldsFilterSchema,
		},
	}
}

type vlanDataSourceResponse struct {
	ID           int64                  `json:"id"`
	Vid          *int64                 `json:"vid"`
	Name         string                 `json:"name"`
	Status       *netboxapi.Choice      `json:"status"`
	Description  string                 `json:"description"`
	Group        *netboxapi.Ref         `json:"group"`
	Role         *netboxapi.Ref         `json:"role"`
	Site         *netboxapi.Ref         `json:"site"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func dataSourceNetboxVlanRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addIntQuery(query, d, "vid", "vid")
	addIntQuery(query, d, "group_id", "group_id")
	addIntQuery(query, d, "role", "role_id")
	addIntQuery(query, d, "tenant", "tenant_id")
	if cfm, ok := d.Get(customFieldsKey).(map[string]interface{}); ok {
		addCustomFieldQuery(query, cfm)
	}

	vlan, err := findOne[vlanDataSourceResponse](m, "ipam/vlans/", query, "vlan")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(vlan.ID, 10))
	d.Set("vid", vlan.Vid)
	d.Set("name", vlan.Name)
	d.Set("status", netboxapi.ChoiceValue(vlan.Status))
	d.Set("description", vlan.Description)
	if vlan.Group != nil {
		d.Set("group_id", vlan.Group.ID)
	}
	if vlan.Role != nil {
		d.Set("role", vlan.Role.ID)
	}
	if vlan.Site != nil {
		d.Set("site", vlan.Site.ID)
	}
	if vlan.Tenant != nil {
		d.Set("tenant", vlan.Tenant.ID)
	}
	d.Set(customFieldsKey, getCustomFields(vlan.CustomFields))
	return nil
}
