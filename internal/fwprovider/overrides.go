package fwprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
		s.ModifyBody = legacyPortsOrMappings
		s.ModifyPlan = legacyPortsPlan
		s.ValidateConfig = protocolMatchesMappings
	},
	"netbox_service_template": func(s *fwresource.Spec) {
		s.ModifyBody = legacyPortsOrMappings
		s.ModifyPlan = legacyPortsPlan
		s.ValidateConfig = protocolMatchesMappings
	},
	"netbox_rack": func(s *fwresource.Spec) {
		// A rack with a rack type gets its form factor, width and height from the type, and NetBox
		// ignores values sent with it. The state does not hold them, so a configuration written from
		// the state (terraform plan -generate-config-out) does not conflict with rack_type_id.
		s.ModifyState = func(_ context.Context, obj map[string]any, _, out fwresource.Values) error {
			if obj["rack_type"] != nil {
				out["form_factor"] = types.StringNull()
				out["width"] = types.Int64Null()
				out["u_height"] = types.Int64Null()
			}
			return nil
		}
	},
	"netbox_config_context": func(s *fwresource.Spec) {
		s.CheckWritten = dataFileWritten("data", true)
		s.ModifyBody = dataFileBody
		s.ModifyPlan = dataFilePlan
		s.ModifyState = dataFileState
	},
	"netbox_config_context_profile": func(s *fwresource.Spec) {
		s.CheckWritten = dataFileWritten("schema", true)
		s.ModifyBody = dataFileBody
		s.ModifyPlan = dataFilePlan
		s.ModifyState = dataFileState
	},
	"netbox_config_template": func(s *fwresource.Spec) {
		s.CheckWritten = dataFileWritten("template_code", false)
		s.ModifyBody = dataFileBody
		s.ModifyPlan = dataFilePlan
		s.ModifyState = dataFileState
	},
	"netbox_export_template": func(s *fwresource.Spec) {
		s.CheckWritten = dataFileWritten("template_code", false)
		s.ModifyBody = dataFileBody
		s.ModifyPlan = dataFilePlan
		s.ModifyState = dataFileState
	},
	"netbox_event_rule": func(s *fwresource.Spec) {
		s.ModifyPlan = aliasPlan("content_types", "object_types")
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
		// The tag lookups are kept while the provider runs. A tag that is created, renamed or deleted in
		// the same run has to be looked up again.
		s.AfterWrite = func(_ context.Context, st *providerstate.State, _ int64, plan, prior fwresource.Values) (bool, error) {
			st.ForgetTag(stringValue(plan, "name"))
			if prior != nil {
				st.ForgetTag(stringValue(prior, "name"))
			}
			return false, nil
		}
		s.BeforeDelete = func(_ context.Context, st *providerstate.State, _ int64, state fwresource.Values) error {
			st.ForgetTag(stringValue(state, "name"))
			return nil
		}
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
		// NetBox shows the key of a token once, in the response to its creation. Later responses
		// have none, or on NetBox 4.4 none at all, and then the key we already know is kept. The object
		// is not read again after a write, because that read would not have the key.
		s.ModifyState = func(_ context.Context, obj map[string]any, prior fwresource.Values, out fwresource.Values) error {
			if key, _ := obj["key"].(string); key != "" {
				return nil
			}
			if k, ok := prior["key"].(types.String); ok && !k.IsNull() && !k.IsUnknown() {
				out["key"] = k
			} else {
				out["key"] = types.StringNull()
			}
			return nil
		}
	},
	"netbox_user": func(s *fwresource.Spec) {
		setAttr(s, "password", func(a *fwresource.Attr) { a.WriteOnly = true })
		s.ModifyBody = func(_ context.Context, plan, prior fwresource.Values, _ map[string]any) error {
			if prior == nil && stringValue(plan, "password") == "" {
				return errors.New("password is required to create a user")
			}
			return nil
		}
		// NetBox 4.5 and later have no is_staff. The state then keeps what the configuration said.
		s.ModifyState = func(_ context.Context, obj map[string]any, prior, out fwresource.Values) error {
			if _, has := obj["is_staff"]; !has && prior != nil {
				if v, ok := prior["staff"]; ok && v != nil && !v.IsNull() && !v.IsUnknown() {
					out["staff"] = v
				}
			}
			return nil
		}
	},
	"netbox_custom_field": func(s *fwresource.Spec) {
		s.ModifyPlan = aliasPlan("content_types", "object_types")
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

// legacyPortsOrMappings sends either port_mappings or the older protocol and ports to NetBox,
// never both, because NetBox rejects an update that has both. port_mappings keeps the value from the
// state while it is not configured, so a change of ports would otherwise be sent together with the
// old mappings.
func legacyPortsOrMappings(_ context.Context, plan, prior fwresource.Values, body map[string]any) error {
	if prior == nil {
		return nil
	}
	a, b := plan["port_mappings"], prior["port_mappings"]
	if a == nil || b == nil || a.IsUnknown() || a.Equal(b) {
		// not configured, or the same as before: protocol and ports are the ones that count
		delete(body, "port_mappings")
		return nil
	}
	delete(body, "protocol")
	delete(body, "ports")
	return nil
}

// protocolMatchesMappings rejects a protocol or ports that differ from port_mappings when both are
// set. NetBox works protocol and ports out from the mappings, so it would otherwise return other
// values than the plan and Terraform would stop with an inconsistent result. Mappings with more
// than one protocol are left to NetBox.
func protocolMatchesMappings(_ context.Context, config fwresource.Values) diag.Diagnostics {
	var diags diag.Diagnostics
	m, ok := config["port_mappings"].(types.Set)
	if !ok || m.IsNull() || m.IsUnknown() || len(m.Elements()) == 0 {
		return diags
	}
	seen := ""
	mapped := map[string]bool{}
	for _, e := range m.Elements() {
		s, ok := e.(types.String)
		if !ok || s.IsNull() || s.IsUnknown() {
			return diags
		}
		p, port, _ := strings.Cut(s.ValueString(), "/")
		if seen != "" && seen != p {
			return diags
		}
		seen = p
		mapped[port] = true
	}
	if protocol := stringValue(config, "protocol"); protocol != "" && seen != protocol {
		diags.AddAttributeError(path.Root("protocol"), "protocol does not match port_mappings",
			fmt.Sprintf("protocol is %q, but port_mappings only holds %q ports. NetBox works the protocol out from port_mappings, so they must match. Change one of them, or leave protocol out.", protocol, seen))
	}
	configured := map[string]bool{}
	name := "ports"
	if ports, ok := config["ports"].(types.Set); ok && !ports.IsNull() {
		if ports.IsUnknown() {
			return diags
		}
		for _, e := range ports.Elements() {
			n, ok := e.(types.Int64)
			if !ok || n.IsNull() || n.IsUnknown() {
				return diags
			}
			configured[strconv.FormatInt(n.ValueInt64(), 10)] = true
		}
	} else if port, ok := config["port"].(types.Int64); ok && !port.IsNull() {
		if port.IsUnknown() {
			return diags
		}
		name = "port"
		configured[strconv.FormatInt(port.ValueInt64(), 10)] = true
	}
	if len(configured) > 0 && !reflect.DeepEqual(configured, mapped) {
		diags.AddAttributeError(path.Root(name), name+" does not match port_mappings",
			fmt.Sprintf("%s and port_mappings hold different ports. NetBox works the ports out from port_mappings, so they must match. Change one of them, or leave %s out.", name, name))
	}
	return diags
}

// legacyPortsPlan lets NetBox work out port_mappings again when protocol or ports change and
// port_mappings is not configured. Without this the plan keeps the old mappings from the state and
// NetBox returns new ones.
func legacyPortsPlan(config, plan, prior fwresource.Values) map[string]attr.Value {
	if prior == nil {
		return nil
	}
	if m, ok := config["port_mappings"].(types.Set); ok && !m.IsNull() {
		return nil
	}
	for _, name := range []string{"protocol", "ports", "port"} {
		a, b := plan[name], prior[name]
		if a != nil && b != nil && !a.Equal(b) {
			return map[string]attr.Value{"port_mappings": types.SetUnknown(types.StringType)}
		}
	}
	return nil
}

// dataFileBody leaves out a data file that is null. NetBox before 4.7.5 refuses null on create, and
// there is nothing to clear unless the object had a data file before.
func dataFileBody(_ context.Context, _, prior fwresource.Values, body map[string]any) error {
	if v, ok := body["data_file"]; !ok || v != nil {
		return nil
	}
	if f, ok := prior["data_file_id"].(types.Int64); ok && !f.IsNull() && f.ValueInt64() != 0 {
		return nil
	}
	delete(body, "data_file")
	return nil
}

// dataFilePlan predicts what NetBox does when the data file is taken away: it derives the data
// source, the data path and the automatic sync from the data file, and clears all of them when the
// object has none. The state would otherwise keep the old values in the plan.
func dataFilePlan(config, plan, prior fwresource.Values) map[string]attr.Value {
	if prior == nil {
		return nil
	}
	if out := dataFileLinkPlan(plan, prior); out != nil {
		return out
	}
	// only a data file that is taken away changes anything. An object that has a data source and no
	// data file (older NetBox versions keep it) is left as it is.
	if f, ok := prior["data_file_id"].(types.Int64); !ok || f.IsNull() || f.ValueInt64() == 0 {
		return nil
	}
	if f, ok := config["data_file_id"].(types.Int64); ok && !f.IsNull() && f.ValueInt64() != 0 {
		return nil
	}
	out := map[string]attr.Value{}
	if _, ok := plan["data_source_id"]; ok {
		out["data_source_id"] = types.Int64Value(0)
	}
	if _, ok := plan["data_path"]; ok {
		out["data_path"] = types.StringValue("")
	}
	if _, ok := plan["auto_sync_enabled"].(types.Bool); ok {
		// only a value the configuration leaves open can be planned
		if c, isBool := config["auto_sync_enabled"].(types.Bool); !isBool || c.IsNull() {
			out["auto_sync_enabled"] = types.BoolValue(false)
		}
	}
	return out
}

// dataFileWritten reports content that NetBox replaced with the content of the data file. From NetBox
// 4.7.5 on, a write with a data file sets the content (template_code, data, schema) from the file, and
// the plan holds the text of the configuration. The plan cannot know the file, because the API does not
// return its content, so the difference only shows after the write.
func dataFileWritten(name string, isJSON bool) func(plan fwresource.Values, obj map[string]any) error {
	return func(plan fwresource.Values, obj map[string]any) error {
		file, ok := plan["data_file_id"].(types.Int64)
		if !ok || file.IsNull() || file.IsUnknown() || file.ValueInt64() == 0 || obj["data_file"] == nil {
			return nil
		}
		want, ok := plan[name].(types.String)
		if !ok || want.IsNull() || want.IsUnknown() {
			return nil
		}
		got := obj[name]
		same := false
		if isJSON {
			var a any
			same = json.Unmarshal([]byte(want.ValueString()), &a) == nil && reflect.DeepEqual(a, got)
		} else {
			s, _ := got.(string)
			same = s == want.ValueString()
		}
		if same {
			return nil
		}
		filePath, _ := obj["data_path"].(string)
		return fmt.Errorf("with data_file_id NetBox keeps %s equal to the content of the data file%s, and replaced the text of the configuration. "+
			"Write the content of the data file in %s, or remove data_file_id. "+
			"NetBox already stored the object with the content of the file, so change the configuration and apply again", name, quotedPath(filePath), name)
	}
}

func quotedPath(path string) string {
	if path == "" {
		return ""
	}
	return " (`" + path + "`)"
}

// dataFileLinkPlan predicts that NetBox sets the data source when an update links a data file. The
// update runs the sync of the data file, which sets the data source, even when the object was created
// with the file and has no data source yet (NetBox before 4.7 does not sync on create). Without this,
// the plan keeps the old data source and the apply ends with "inconsistent result".
//
// It only acts when the update happens anyway: the framework has marked last_updated as unknown then.
func dataFileLinkPlan(plan, prior fwresource.Values) map[string]attr.Value {
	if lu, ok := plan["last_updated"]; !ok || !lu.IsUnknown() {
		return nil
	}
	if _, ok := plan["data_source_id"]; !ok {
		return nil
	}
	file, ok := plan["data_file_id"].(types.Int64)
	if !ok || file.IsNull() || file.IsUnknown() || file.ValueInt64() == 0 {
		return nil
	}
	priorFile, _ := prior["data_file_id"].(types.Int64)
	priorSource, _ := prior["data_source_id"].(types.Int64)
	if priorFile.ValueInt64() == file.ValueInt64() && priorSource.ValueInt64() != 0 {
		return nil
	}
	return map[string]attr.Value{"data_source_id": types.Int64Unknown()}
}

// dataFileState keeps data_source_id at 0 while the object has no data file, as it was before the
// data file could be set, so a configuration that says 0 stays valid.
func dataFileState(_ context.Context, obj map[string]any, prior, out fwresource.Values) error {
	if obj["data_source"] == nil {
		if _, ok := out["data_source_id"]; ok {
			out["data_source_id"] = types.Int64Value(0)
		}
	}
	// NetBox 4.4 has no automatic sync, and the response has no auto_sync_enabled. The state then keeps
	// what the configuration said (false, as nothing can be synced), or false when it said nothing.
	if _, has := obj["auto_sync_enabled"]; !has {
		if _, ok := out["auto_sync_enabled"]; ok {
			keep := types.BoolValue(false)
			if v, isBool := prior["auto_sync_enabled"].(types.Bool); isBool && !v.IsNull() && !v.IsUnknown() {
				keep = v
			}
			out["auto_sync_enabled"] = keep
		}
	}
	return nil
}

// aliasPlan keeps an old attribute name and its new name equal in the plan: the one the configuration
// leaves out takes the value of the other.
func aliasPlan(oldName, newName string) func(config, plan, prior fwresource.Values) map[string]attr.Value {
	set := func(v attr.Value) bool { return v != nil && !v.IsNull() && !v.IsUnknown() }
	return func(config, plan, _ fwresource.Values) map[string]attr.Value {
		switch {
		case set(config[oldName]) && !set(config[newName]):
			return map[string]attr.Value{newName: plan[oldName]}
		case set(config[newName]) && !set(config[oldName]):
			return map[string]attr.Value{oldName: plan[newName]}
		}
		return nil
	}
}
