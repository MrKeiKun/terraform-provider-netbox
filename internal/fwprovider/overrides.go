package fwprovider

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// overrides adjust the standard specs of resources that do not follow the usual pattern of
// attributes that map one to one onto fields of the API.
var overrides = map[string]func(s *fwresource.Spec){
	"netbox_circuit_termination": func(s *fwresource.Spec) {
		s.Scopes = []fwresource.Scope{{
			TypeKey: "termination_type",
			IDKey:   "termination_id",
			Options: []fwresource.ScopeOption{
				{Attr: "site_id", Type: "dcim.site"},
				{Attr: "site_group_id", Type: "dcim.sitegroup"},
				{Attr: "location_id", Type: "dcim.location"},
				{Attr: "region_id", Type: "dcim.region"},
				{Attr: "provider_network_id", Type: "circuits.providernetwork"},
			},
		}}
	},
	"netbox_cluster": func(s *fwresource.Spec) {
		s.Scopes = []fwresource.Scope{{
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
	"netbox_prefix": func(s *fwresource.Spec) {
		s.Scopes = []fwresource.Scope{{
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
	"netbox_mac_address": func(s *fwresource.Spec) {
		s.Scopes = []fwresource.Scope{{
			TypeKey:   "assigned_object_type",
			IDKey:     "assigned_object_id",
			BlankType: true,
			ByPrior:   true,
			Options: []fwresource.ScopeOption{
				{Attr: "virtual_machine_interface_id", Type: "virtualization.vminterface"},
				{Attr: "device_interface_id", Type: "dcim.interface"},
				{Attr: "interface_id", TypeAttr: "object_type"},
			},
		}}
		setAttr(s, "mac_address", func(a *fwresource.Attr) { a.CaseInsensitive = true })
	},
	"netbox_vpn_tunnel_termination": func(s *fwresource.Spec) {
		s.Scopes = []fwresource.Scope{{
			TypeKey: "termination_type",
			IDKey:   "termination_id",
			Options: []fwresource.ScopeOption{
				{Attr: "virtual_machine_interface_id", Type: "virtualization.vminterface"},
				{Attr: "device_interface_id", Type: "dcim.interface"},
			},
		}}
	},
	"netbox_service": func(s *fwresource.Spec) {
		// the plain tags attribute of this resource holds the tag names like everywhere else
		removeAttr(s, "tags")
		s.Tags = true
		s.Scopes = []fwresource.Scope{{
			TypeKey: "parent_object_type",
			IDKey:   "parent_object_id",
			Options: []fwresource.ScopeOption{
				{Attr: "virtual_machine_id", Type: "virtualization.virtualmachine"},
				{Attr: "device_id", Type: "dcim.device"},
			},
		}}
	},
	"netbox_event_rule": func(s *fwresource.Spec) {
		s.ModifyBody = func(_ context.Context, plan, _ fwresource.Values, body map[string]any) error {
			if t, ok := plan["action_type"].(types.String); ok {
				switch t.ValueString() {
				case "webhook":
					body["action_object_type"] = "extras.webhook"
				case "script":
					body["action_object_type"] = "extras.script"
				}
			}
			return nil
		}
	},
	"netbox_power_feed": func(s *fwresource.Spec) {
		setAttr(s, "max_percent_utilization", func(a *fwresource.Attr) { a.Key = "max_utilization" })
	},
	"netbox_site": func(s *fwresource.Spec) {
		setAttr(s, "timezone", func(a *fwresource.Attr) { a.Key = "time_zone" })
	},
	"netbox_contact": func(s *fwresource.Spec) {
		// group_id is the deprecated way to set a single group; the API has a list of groups.
		setAttr(s, "group_id", func(a *fwresource.Attr) { a.NoSend = true })
		s.ModifyBody = func(_ context.Context, plan, _ fwresource.Values, body map[string]any) error {
			if id := intValue(plan, "group_id"); id != 0 && listLen(plan["group_ids"]) == 0 {
				body["groups"] = []any{id}
			}
			return nil
		}
		s.ModifyState = func(_ context.Context, obj map[string]any, prior, out fwresource.Values) error {
			out["group_id"] = types.Int64Value(0)
			if intValue(prior, "group_id") == 0 || listLen(prior["group_ids"]) != 0 {
				return nil
			}
			out["group_ids"] = prior["group_ids"]
			if groups, ok := obj["groups"].([]any); ok && len(groups) == 1 {
				if m, ok := groups[0].(map[string]any); ok {
					if id, ok := m["id"].(float64); ok {
						out["group_id"] = types.Int64Value(int64(id))
					}
				}
			}
			return nil
		}
	},
	"netbox_device_interface": func(s *fwresource.Spec) {
		// mac_address is the first of the MAC addresses assigned to the interface
		s.ModifyState = func(_ context.Context, obj map[string]any, prior, out fwresource.Values) error {
			if macs, ok := obj["mac_addresses"].([]any); ok && len(macs) > 0 {
				if m, ok := macs[0].(map[string]any); ok {
					if v, ok := m["mac_address"].(string); ok {
						out["mac_address"] = types.StringValue(v)
					}
				}
			}
			return nil
		}
	},
	"netbox_tag": func(s *fwresource.Spec) {
		s.NoDefaultTags = true
	},
	"netbox_ip_address": func(s *fwresource.Spec) {
		setAttr(s, "ip_address", func(a *fwresource.Attr) { a.Key = "address" })
		setAttr(s, "nat_inside_address_id", func(a *fwresource.Attr) { a.Key = "nat_inside" })
		setAttr(s, "dns_name", func(a *fwresource.Attr) { a.CaseInsensitive = true })
		s.Scopes = []fwresource.Scope{{
			TypeKey:         "assigned_object_type",
			IDKey:           "assigned_object_id",
			BlankType:       true,
			ByPrior:         true,
			GenericFallback: true,
			Options: []fwresource.ScopeOption{
				{Attr: "virtual_machine_interface_id", Type: "virtualization.vminterface"},
				{Attr: "device_interface_id", Type: "dcim.interface"},
				{Attr: "interface_id", TypeAttr: "object_type"},
			},
		}}
		setAttr(s, "nat_outside_addresses", func(a *fwresource.Attr) { a.Key = "nat_outside" })
		setSub(s, "nat_outside_addresses", "ip_address", func(a *fwresource.Attr) { a.Key = "address" })
		setSub(s, "nat_outside_addresses", "address_family", func(a *fwresource.Attr) { a.Key = "family" })
	},
	"netbox_device_front_port": func(s *fwresource.Spec) {
		setSub(s, "rear_ports", "rear_port_id", func(a *fwresource.Attr) { a.Key = "rear_port" })
		portMappingHooks(s, "rear_ports", "rear_port", true)
	},
	"netbox_front_port_template": func(s *fwresource.Spec) {
		setSub(s, "rear_ports", "rear_port_id", func(a *fwresource.Attr) { a.Key = "rear_port" })
		portMappingHooks(s, "rear_ports", "rear_port", true)
	},
	"netbox_device_rear_port": func(s *fwresource.Spec) {
		setSub(s, "front_ports", "front_port_id", func(a *fwresource.Attr) { a.Key = "front_port" })
		portMappingHooks(s, "front_ports", "front_port", false)
	},
	"netbox_rear_port_template": func(s *fwresource.Spec) {
		setSub(s, "front_ports", "front_port_id", func(a *fwresource.Attr) { a.Key = "front_port" })
		portMappingHooks(s, "front_ports", "front_port", false)
	},
	"netbox_token": func(s *fwresource.Spec) {
		setAttr(s, "expires", func(a *fwresource.Attr) { a.Time = true; a.NoZero = true })
	},
	"netbox_user": func(s *fwresource.Spec) {
		setAttr(s, "password", func(a *fwresource.Attr) { a.WriteOnly = true })
	},
	"netbox_custom_field": func(s *fwresource.Spec) {
		setAttr(s, "default", func(a *fwresource.Attr) {
			a.Kind = fwresource.String
			a.LooseJSON = true
		})
	},
}

func init() {
	for _, spec := range standardResources {
		if f, ok := overrides[spec.TypeName]; ok {
			f(spec)
		}
	}
}

// setAttr changes the attribute of a spec with the given name.
func setAttr(s *fwresource.Spec, name string, change func(a *fwresource.Attr)) {
	for i := range s.Attrs {
		if s.Attrs[i].Name == name {
			change(&s.Attrs[i])
			return
		}
	}
	panic("fwprovider: " + s.TypeName + " has no attribute " + name)
}

// removeAttr removes the attribute of a spec with the given name.
func removeAttr(s *fwresource.Spec, name string) {
	for i := range s.Attrs {
		if s.Attrs[i].Name == name {
			s.Attrs = append(s.Attrs[:i], s.Attrs[i+1:]...)
			return
		}
	}
	panic("fwprovider: " + s.TypeName + " has no attribute " + name)
}

// setSub changes an attribute of the objects in a nested attribute.
func setSub(s *fwresource.Spec, nested, name string, change func(a *fwresource.Attr)) {
	setAttr(s, nested, func(n *fwresource.Attr) {
		for i := range n.Sub {
			if n.Sub[i].Name == name {
				change(&n.Sub[i])
				return
			}
		}
		panic("fwprovider: " + s.TypeName + "." + nested + " has no attribute " + name)
	})
}

// portMappingHooks handles the mapping between a front port and rear ports (and the other way
// around). NetBox 4.5 and later describe it with a list of mappings. The single port attributes
// (for example rear_port_id and rear_port_position) are kept for compatibility, they stand for a
// list with one mapping.
func portMappingHooks(s *fwresource.Spec, list, portKey string, withSingle bool) {
	equal := func(a, b attr.Value) bool {
		if a == nil || b == nil {
			return a == b
		}
		return a.Equal(b)
	}
	s.ModifyBody = func(_ context.Context, plan, prior fwresource.Values, body map[string]any) error {
		if withSingle {
			idAttr, posAttr := portKey+"_id", portKey+"_position"
			if n, ok := plan[idAttr].(types.Int64); ok && !n.IsNull() && !n.IsUnknown() && n.ValueInt64() != 0 {
				position := int64(1)
				if p, ok := plan[posAttr].(types.Int64); ok && !p.IsNull() && !p.IsUnknown() && p.ValueInt64() != 0 {
					position = p.ValueInt64()
				}
				body[list] = []any{map[string]any{"position": int64(1), portKey: n.ValueInt64(), portKey + "_position": position}}
			}
			delete(body, portKey+"_position")
			if prior != nil && equal(plan[list], prior[list]) && equal(plan[idAttr], prior[idAttr]) && equal(plan[posAttr], prior[posAttr]) {
				// NetBox rejects an update that sends the mapping it already has
				delete(body, list)
				delete(body, portKey)
			}
			return nil
		}
		if prior != nil && equal(plan[list], prior[list]) {
			delete(body, list)
		}
		return nil
	}
	if withSingle {
		s.ModifyState = func(_ context.Context, obj map[string]any, _ fwresource.Values, out fwresource.Values) error {
			if obj[portKey] != nil {
				return nil
			}
			mappings, _ := obj[list].([]any)
			if len(mappings) == 0 {
				return nil
			}
			first, _ := mappings[0].(map[string]any)
			out[portKey+"_id"] = types.Int64Value(portInt(first[portKey]))
			out[portKey+"_position"] = types.Int64Value(portInt(first[portKey+"_position"]))
			return nil
		}
	}
}

func portInt(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case map[string]any:
		return portInt(x["id"])
	}
	return 0
}

// listLen is the number of elements of a list or set value.
func listLen(v attr.Value) int {
	switch x := v.(type) {
	case types.List:
		return len(x.Elements())
	case types.Set:
		return len(x.Elements())
	}
	return 0
}
