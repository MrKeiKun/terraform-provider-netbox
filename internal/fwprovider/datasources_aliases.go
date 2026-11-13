package fwprovider

import (
	"context"
	"fmt"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// dataSourceAliases are data sources that are known by a second name, so that configurations written
// for either name keep working. The key is the new name, the value the data source it copies.
var dataSourceAliases = map[string]string{
	"netbox_circuit_provider_account":   "netbox_provider_account",
	"netbox_virtual_disks":              "netbox_virtual_disk",
	"netbox_virtual_machine_interfaces": "netbox_interfaces",
	"netbox_vpn_tunnels":                "netbox_tunnels",
}

func init() {
	customDataSources = append(customDataSources, supportedVersions())
}

// supportedVersions lists the NetBox versions this release of the provider was tested against.
func supportedVersions() *fwresource.DataSourceSpec {
	return &fwresource.DataSourceSpec{
		TypeName:    "netbox_supported_versions",
		Description: ":meta:subcategory:Extras:The NetBox versions this provider release was tested against. Any other version only gets a \"Possibly unsupported Netbox version\" warning at startup.",
		Noun:        "supported versions",
		Attrs: []fwresource.Attr{
			{Name: "versions", Kind: fwresource.StringList, Computed: true, Description: "The supported NetBox versions."},
		},
		Custom: func(_ context.Context, _ *providerstate.State, _ fwresource.Values, _ datasource.ReadRequest) ([]map[string]any, error) {
			versions := make([]any, len(providerstate.SupportedVersions))
			for i, v := range providerstate.SupportedVersions {
				versions[i] = v
			}
			return []map[string]any{{"versions": versions}}, nil
		},
	}
}

// aliasedDataSources returns the copies of the data sources that have a second name. They are made when
// they are asked for, so that they also have what the other files add to the data source at start up.
func aliasedDataSources() []*fwresource.DataSourceSpec {
	all := append(HandwrittenDataSources(), generatedDataSources...)
	byName := make(map[string]*fwresource.DataSourceSpec, len(all))
	for _, s := range all {
		byName[s.TypeName] = s
	}
	var names []string
	for name := range dataSourceAliases {
		names = append(names, name)
	}
	sortStrings(names)
	out := make([]*fwresource.DataSourceSpec, 0, len(names))
	for _, name := range names {
		from, ok := byName[dataSourceAliases[name]]
		if !ok {
			panic(fmt.Sprintf("fwprovider: %s is an alias of %s, which does not exist", name, dataSourceAliases[name]))
		}
		c := *from
		c.TypeName = name
		out = append(out, &c)
	}
	return out
}
