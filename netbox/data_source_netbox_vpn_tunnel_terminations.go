package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var datasourceNetboxVpnTunnelTerminationRoleOptions = []string{"peer", "hub", "spoke"}

func dataSourceNetboxVpnTunnelTerminations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetboxVpnTunnelTerminationsRead,
		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/features/vpn-tunnels/):

> NetBox can model private tunnels formed among virtual termination points across your network. Typical tunnel implementations include GRE, IP-in-IP, and IPSec. A tunnel may be terminated to two or more device or virtual machine interfaces. For convenient organization, tunnels may be assigned to user-defined groups.`,
		Schema: map[string]*schema.Schema{
			"limit": {
				Type:             schema.TypeInt,
				Description:      "The maximum number of items to return. Defaults to 0, which returns all items.",
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
			},
			"terminations": {
				Type:        schema.TypeList,
				Description: "The terminations of the vpn tunnel terminations.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the vpn tunnel terminations.",
							Computed:    true,
						},
						"tunnel_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the tunnel associated with the vpn tunnel terminations.",
							Computed:    true,
						},
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: buildValidValueDescription(datasourceNetboxVpnTunnelTerminationRoleOptions),
						},
						"termination_type": {
							Type:        schema.TypeString,
							Description: "The termination type of the vpn tunnel terminations.",
							Computed:    true,
						},
						"termination_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the termination associated with the vpn tunnel terminations.",
							Computed:    true,
						},
						tagsKey: tagsSchema,
					},
				},
			},
		},
	}
}

func dataSourceNetboxVpnTunnelTerminationsRead(d *schema.ResourceData, m interface{}) error {
	// this data source has no filters
	results, err := listFiltered[map[string]interface{}](d, m, "vpn/tunnel-terminations/", "Tunnel Terminations", map[string]string{}, nil)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["tunnel_id"] = jID(v, "tunnel")
		mapping["role"] = jChoice(v, "role")
		mapping["termination_type"] = jStr(v, "termination_type")
		mapping["termination_id"] = jInt(v, "termination_id")
		mapping["tags"] = jTags(v)

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("terminations", s)
}
