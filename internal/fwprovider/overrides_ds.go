package fwprovider

import "github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"

// dataSourceOverrides adjust the standard specs of data sources that do not follow the usual pattern.
var dataSourceOverrides = map[string]func(s *fwresource.DataSourceSpec){
	"netbox_clusters": func(s *fwresource.DataSourceSpec) {
		s.ResultScopes = []fwresource.Scope{{
			TypeKey: "scope_type",
			IDKey:   "scope_id",
			Options: []fwresource.ScopeOption{
				{Attr: "site_id", Type: "dcim.site"},
				{Attr: "site_group_id", Type: "dcim.sitegroup"},
				{Attr: "location_id", Type: "dcim.location"},
				{Attr: "region_id", Type: "dcim.region"},
			},
		}}
	},
	"netbox_prefixes": func(s *fwresource.DataSourceSpec) {
		s.ResultScopes = []fwresource.Scope{{
			TypeKey: "scope_type",
			IDKey:   "scope_id",
			Options: []fwresource.ScopeOption{
				{Attr: "site_id", Type: "dcim.site"},
				{Attr: "site_group_id", Type: "dcim.sitegroup"},
				{Attr: "location_id", Type: "dcim.location"},
				{Attr: "region_id", Type: "dcim.region"},
			},
		}}
	},
}

func init() {
	for _, spec := range standardDataSources {
		if f, ok := dataSourceOverrides[spec.TypeName]; ok {
			f(spec)
		}
	}
}
