package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func setDSAttr(s *fwresource.DataSourceSpec, name string, change func(a *fwresource.Attr)) {
	for i := range s.Attrs {
		if s.Attrs[i].Name == name {
			change(&s.Attrs[i])
			return
		}
	}
	panic("fwprovider: " + s.TypeName + " has no attribute " + name)
}

func intCfg(cfg fwresource.Values, name string) int64 {
	if n, ok := cfg[name].(types.Int64); ok && !n.IsNull() && !n.IsUnknown() {
		return n.ValueInt64()
	}
	return 0
}

func strCfg(cfg fwresource.Values, name string) string {
	if s, ok := cfg[name].(types.String); ok && !s.IsNull() && !s.IsUnknown() {
		return s.ValueString()
	}
	return ""
}

// netboxTime writes a time the way NetBox shows it in the data sources: in UTC with milliseconds.
func netboxTime(v any) any {
	s, _ := v.(string)
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return v
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func init() {
	custom := map[string]func(s *fwresource.DataSourceSpec){
		"netbox_available_prefix": func(s *fwresource.DataSourceSpec) {
			s.Plural.NoFilter = true
			setDSAttr(s, "vrf_id", func(a *fwresource.Attr) { a.Key = "vrf" })
			s.Custom = func(ctx context.Context, st *providerstate.State, cfg fwresource.Values, _ datasource.ReadRequest) ([]map[string]any, error) {
				path := fmt.Sprintf("ipam/prefixes/%d/available-prefixes/", intCfg(cfg, "prefix_id"))
				res, err := netboxapi.Request[[]map[string]any](ctx, st.Rest, http.MethodGet, path, nil, nil)
				if err != nil {
					return nil, err
				}
				return *res, nil
			}
		},
		"netbox_device_render_config": func(s *fwresource.DataSourceSpec) {
			setDSAttr(s, "config_template_id", func(a *fwresource.Attr) { a.Path = []string{"configtemplate", "id"} })
			setDSAttr(s, "config_template_name", func(a *fwresource.Attr) { a.Path = []string{"configtemplate", "name"} })
			s.Custom = func(ctx context.Context, st *providerstate.State, cfg fwresource.Values, _ datasource.ReadRequest) ([]map[string]any, error) {
				deviceID := intCfg(cfg, "device_id")
				// Rendering a config is an action on the device, NetBox wants a POST for it.
				path := fmt.Sprintf("dcim/devices/%d/render-config/", deviceID)
				res, err := netboxapi.Request[map[string]any](ctx, st.Rest, http.MethodPost, path, nil, map[string]any{})
				if err != nil {
					return nil, err
				}
				(*res)["id"] = float64(deviceID)
				return []map[string]any{*res}, nil
			}
		},
		"netbox_ip_address": func(s *fwresource.DataSourceSpec) {
			setDSAttr(s, "ip_address", func(a *fwresource.Attr) { a.Key = "address" })
			setDSAttr(s, "address_family", func(a *fwresource.Attr) { a.Path = []string{"family", "label"} })
			setDSAttr(s, "role", func(a *fwresource.Attr) { a.Path = []string{"role", "value"} })
			s.Custom = func(ctx context.Context, st *providerstate.State, cfg fwresource.Values, _ datasource.ReadRequest) ([]map[string]any, error) {
				obj, err := netboxapi.Get[map[string]any](ctx, st.Rest, "ipam/ip-addresses/", intCfg(cfg, "id"))
				if err != nil {
					return nil, err
				}
				(*obj)["created"] = netboxTime((*obj)["created"])
				(*obj)["last_updated"] = netboxTime((*obj)["last_updated"])
				return []map[string]any{*obj}, nil
			}
		},
		"netbox_notification_group": func(s *fwresource.DataSourceSpec) {
			setDSAttr(s, "group_ids", func(a *fwresource.Attr) { a.Key = "groups" })
			setDSAttr(s, "user_ids", func(a *fwresource.Attr) { a.Key = "users" })
			s.Noun = "notification group"
			s.Custom = func(ctx context.Context, st *providerstate.State, cfg fwresource.Values, _ datasource.ReadRequest) ([]map[string]any, error) {
				name := strCfg(cfg, "name")
				// NetBox ignores the name filter for notification groups, so the search is used to
				// narrow the list down and the exact name is matched here.
				groups, err := netboxapi.List[map[string]any](ctx, st.Rest, "extras/notification-groups/", url.Values{"q": {name}})
				if err != nil {
					return nil, err
				}
				var matches []map[string]any
				for _, g := range groups {
					if n, _ := g["name"].(string); n == name {
						matches = append(matches, g)
					}
				}
				return matches, nil
			}
		},
		"netbox_rack_type": func(s *fwresource.DataSourceSpec) {
			setDSAttr(s, "manufacturer_id", func(a *fwresource.Attr) { a.Key = "manufacturer" })
			setDSAttr(s, "mounting_depth_mm", func(a *fwresource.Attr) { a.Key = "mounting_depth" })
			s.Noun = "rack type"
			s.Custom = func(ctx context.Context, st *providerstate.State, cfg fwresource.Values, _ datasource.ReadRequest) ([]map[string]any, error) {
				model, slug := strCfg(cfg, "model"), strCfg(cfg, "slug")
				if model == "" && slug == "" {
					return nil, fmt.Errorf("either 'model' or 'slug' must be specified")
				}
				// The rack-types list endpoint has no server-side model/slug filter, so the rack
				// types are listed and matched exactly here.
				all, err := netboxapi.List[map[string]any](ctx, st.Rest, "dcim/rack-types/", nil)
				if err != nil {
					return nil, err
				}
				var matches []map[string]any
				for _, rt := range all {
					if m, _ := rt["model"].(string); model != "" && m != model {
						continue
					}
					if sl, _ := rt["slug"].(string); slug != "" && sl != slug {
						continue
					}
					matches = append(matches, rt)
				}
				return matches, nil
			}
		},
	}
	for _, spec := range customDataSources {
		if f, ok := custom[spec.TypeName]; ok {
			f(spec)
		}
	}
	for _, spec := range customDataSources {
		if spec.TypeName != "netbox_virtual_disk" {
			continue
		}
		spec.Path = "virtualization/virtual-disks/"
		spec.Plural.Filters = map[string]string{"name": "name__ic", "virtual_machine_id": "virtual_machine_id", "tag": "tag"}
		spec.Plural.NameRegex = "name_regex"
		setDSAttr(spec, "size_mb", func(a *fwresource.Attr) { a.Key = "size" })
		setDSAttr(spec, "virtual_machine_id", func(a *fwresource.Attr) { a.Key = "virtual_machine" })
		setDSAttr(spec, "tags", func(a *fwresource.Attr) { a.ElemKey = "name" })
		spec.ModifyQuery = func(_ context.Context, cfg fwresource.Values, q url.Values) error {
			if l, ok := cfg["ordering"].(types.List); ok && !l.IsNull() && !l.IsUnknown() {
				var parts []string
				for _, e := range l.Elements() {
					if s, ok := e.(types.String); ok {
						parts = append(parts, s.ValueString())
					}
				}
				if len(parts) > 0 {
					q.Set("ordering", strings.Join(parts, ","))
				}
			}
			return nil
		}
	}
}
