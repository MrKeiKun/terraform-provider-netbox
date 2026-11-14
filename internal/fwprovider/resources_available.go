package fwprovider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The "available" resources let NetBox pick a free IP address, prefix or VLAN. They are derived
// from the resources of the regular objects, so their attributes stay the same.

func init() {
	manualResources = append(manualResources, availableIPAddress(), availablePrefix(), availableVLAN())
}

// derive returns a copy of a standard spec with the attributes of the regular resource.
func derive(typeName string) *fwresource.Spec {
	for _, s := range standardResources {
		if s.TypeName == typeName {
			c := *s
			c.Attrs = append([]fwresource.Attr(nil), s.Attrs...)
			return &c
		}
	}
	panic("fwprovider: no standard resource " + typeName)
}

func dropAttrs(s *fwresource.Spec, names ...string) {
	for _, n := range names {
		removeAttr(s, n)
	}
}

func rename(s *fwresource.Spec, from, to string) {
	for i := range s.Attrs {
		s.Attrs[i].Description = strings.ReplaceAll(s.Attrs[i].Description, from, to)
		for j := range s.Attrs[i].Sub {
			s.Attrs[i].Sub[j].Description = strings.ReplaceAll(s.Attrs[i].Sub[j].Description, from, to)
		}
	}
}

func intValue(v fwresource.Values, name string) int64 {
	if n, ok := v[name].(types.Int64); ok && !n.IsNull() && !n.IsUnknown() {
		return n.ValueInt64()
	}
	return 0
}

func stringValue(v fwresource.Values, name string) string {
	if s, ok := v[name].(types.String); ok && !s.IsNull() && !s.IsUnknown() {
		return s.ValueString()
	}
	return ""
}

func availableIPAddress() *fwresource.Spec {
	s := derive("netbox_ip_address")
	dropAttrs(s, "nat_outside_addresses")
	s.TypeName = "netbox_available_ip_address"
	s.Description = ":meta:subcategory:IP Address Management (IPAM):Per [the docs](https://netbox.readthedocs.io/en/stable/models/ipam/ipaddress/):\n\n> An IP address comprises a single host address (either IPv4 or IPv6) and its subnet mask. Its mask should match exactly how the IP address is configured on an interface in the real world.\n> Like a prefix, an IP address can optionally be assigned to a VRF (otherwise, it will appear in the \"global\" table). IP addresses are automatically arranged under parent prefixes within their respective VRFs according to the IP hierarchya.\n>\n> Each IP address can also be assigned an operational status and a functional role. Statuses are hard-coded in NetBox and include the following:\n> * Active\n> * Reserved\n> * Deprecated\n> * DHCP\n> * SLAAC (IPv6 Stateless Address Autoconfiguration)\n\nThis resource will retrieve the next available IP address from a given prefix or IP range (specified by ID)"
	rename(s, "the ip address", "the available ip address")
	setAttr(s, "ip_address", func(a *fwresource.Attr) {
		a.Required = false
		a.Computed = true
		a.Description = "The IP address of the available ip address, in CIDR notation."
	})
	setAttr(s, "status", func(a *fwresource.Attr) {
		a.Required = false
		a.Optional = true
		a.Default = "active"
		a.OneOf = nil
	})
	s.Attrs = append(s.Attrs,
		fwresource.Attr{Name: "prefix_id", Kind: fwresource.Int, Optional: true, Local: true, ForceNew: true, Description: "The ID of the prefix associated with the available ip address."},
		fwresource.Attr{Name: "ip_range_id", Kind: fwresource.Int, Optional: true, Local: true, ForceNew: true, Description: "The ID of the ip range associated with the available ip address."},
	)
	s.ExactlyOneOf = [][]string{{"prefix_id", "ip_range_id"}}
	s.ModifyBody = func(_ context.Context, _, prior fwresource.Values, body map[string]any) error {
		if prior != nil {
			body["address"] = stringValue(prior, "ip_address")
		}
		return nil
	}
	s.Allocate = func(ctx context.Context, st *providerstate.State, plan fwresource.Values, body map[string]any) (map[string]any, error) {
		entry := map[string]any{}
		// Send what NetBox may require when it creates the address, such as
		// mandatory custom fields. The update that follows sets everything else.
		if id, ok := body["vrf"].(int64); ok && id != 0 {
			entry["vrf"] = map[string]any{"id": id}
		}
		if id, ok := body["tenant"].(int64); ok && id != 0 {
			entry["tenant"] = map[string]any{"id": id}
		}
		if cf, ok := body["custom_fields"]; ok {
			entry["custom_fields"] = cf
		}
		path, what := "", ""
		if id := intValue(plan, "prefix_id"); id != 0 {
			path, what = fmt.Sprintf("ipam/prefixes/%d/available-ips/", id), fmt.Sprintf("prefix %d", id)
		} else if id := intValue(plan, "ip_range_id"); id != 0 {
			path, what = fmt.Sprintf("ipam/ip-ranges/%d/available-ips/", id), fmt.Sprintf("IP range %d", id)
		} else {
			return nil, fmt.Errorf("one of prefix_id and ip_range_id is required")
		}
		created, err := netboxapi.Request[[]map[string]any](ctx, st.Rest, http.MethodPost, path, nil, []any{entry})
		if err != nil {
			return nil, err
		}
		if len(*created) == 0 {
			return nil, fmt.Errorf("no available IP addresses in %s", what)
		}
		return (*created)[0], nil
	}
	s.Finish = func(ctx context.Context, st *providerstate.State, obj map[string]any, prior, out fwresource.Values) error {
		// After an import neither parent is known. Restore the most specific
		// prefix, so the next plan does not want to replace the address.
		if intValue(prior, "prefix_id") != 0 || intValue(prior, "ip_range_id") != 0 {
			return nil
		}
		if id, ok := findParentPrefixID(ctx, st, obj); ok {
			out["prefix_id"] = types.Int64Value(id)
		}
		return nil
	}
	return s
}

// findParentPrefixID returns the ID of the most specific prefix that contains the address and
// lives in the same VRF.
func findParentPrefixID(ctx context.Context, st *providerstate.State, ip map[string]any) (int64, bool) {
	address, _ := ip["address"].(string)
	host, _, err := net.ParseCIDR(address)
	if err != nil {
		return 0, false
	}
	query := url.Values{}
	query.Set("contains", host.String())
	if vrf, ok := ip["vrf"].(map[string]any); ok {
		if id, ok := vrf["id"].(float64); ok {
			query.Set("vrf_id", strconv.FormatInt(int64(id), 10))
		}
	} else {
		query.Set("vrf_id", "null")
	}
	prefixes, err := netboxapi.List[struct {
		ID     int64  `json:"id"`
		Prefix string `json:"prefix"`
	}](ctx, st.Rest, "ipam/prefixes/", query)
	if err != nil {
		return 0, false
	}
	var bestID int64
	bestLen := -1
	for _, p := range prefixes {
		_, n, err := net.ParseCIDR(p.Prefix)
		if err != nil {
			continue
		}
		if ones, _ := n.Mask.Size(); ones > bestLen {
			bestLen = ones
			bestID = p.ID
		}
	}
	return bestID, bestLen >= 0
}

func availablePrefix() *fwresource.Spec {
	s := derive("netbox_prefix")
	s.TypeName = "netbox_available_prefix"
	s.Description = ":meta:subcategory:IP Address Management (IPAM):Allocates the next free prefix of the requested length from a parent prefix. Import with \"<parent_prefix_id> <prefix_id> <prefix_length>\"."
	rename(s, "of the prefix", "of the available prefix")
	rename(s, "with the prefix", "with the available prefix")
	rename(s, "about the prefix", "about the available prefix")
	rename(s, "Whether the prefix", "Whether the available prefix")
	setAttr(s, "prefix", func(a *fwresource.Attr) {
		a.Required = false
		a.Computed = true
	})
	s.Attrs = append(s.Attrs,
		fwresource.Attr{Name: "parent_prefix_id", Kind: fwresource.Int, Required: true, Local: true, ForceNew: true, Description: "The ID of the parent prefix associated with the available prefix."},
		fwresource.Attr{Name: "prefix_length", Kind: fwresource.Int, Required: true, Local: true, ForceNew: true, Min: i64(0), Max: i64(128), Description: "The prefix length of the available prefix."},
	)
	s.ModifyBody = func(_ context.Context, _, prior fwresource.Values, body map[string]any) error {
		if prior != nil {
			body["prefix"] = stringValue(prior, "prefix")
		}
		return nil
	}
	s.Allocate = func(ctx context.Context, st *providerstate.State, plan fwresource.Values, body map[string]any) (map[string]any, error) {
		// Send every configured field with the allocation, so NetBox rules that need
		// a tenant, role and so on on creation accept it. The prefix itself is chosen
		// by NetBox. The update that follows sets whatever NetBox did not take over.
		data := map[string]any{}
		for k, v := range body {
			data[k] = v
		}
		delete(data, "prefix")
		data["prefix_length"] = intValue(plan, "prefix_length")
		path := fmt.Sprintf("ipam/prefixes/%d/available-prefixes/", intValue(plan, "parent_prefix_id"))
		created, err := netboxapi.Request[map[string]any](ctx, st.Rest, http.MethodPost, path, nil, data)
		if err != nil {
			return nil, err
		}
		return *created, nil
	}
	s.ImportID = func(id string) (string, map[string]any, error) {
		parts := strings.SplitN(id, " ", 3)
		if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return "", nil, fmt.Errorf("unexpected format of (%s), expected 'parent_prefix_id prefix_id prefix_length'", id)
		}
		parentID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return "", nil, fmt.Errorf("parent_id (%s) is not an integer", parts[0])
		}
		if _, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
			return "", nil, fmt.Errorf("prefix_id (%s) is not an integer", parts[1])
		}
		length, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return "", nil, fmt.Errorf("prefix_length (%s) is not an integer", parts[2])
		}
		return parts[1], map[string]any{"parent_prefix_id": parentID, "prefix_length": length}, nil
	}
	return s
}

func availableVLAN() *fwresource.Spec {
	return &fwresource.Spec{
		TypeName:     "netbox_available_vlan",
		Path:         "ipam/vlans/",
		Description:  ":meta:subcategory:IP Address Management (IPAM):Per [the docs](https://netbox.readthedocs.io/en/stable/models/ipam/vlan/):\n\n> A VLAN represents an isolated Layer 2 domain identified by a numeric ID (1–4094). VLANs may be assigned to specific sites or marked as global.\n> Optionally, they can be organized within VLAN groups to define scope and enforce uniqueness.\n>\n> Each VLAN can also be assigned an operational status and a functional role. Statuses are hard-coded in NetBox and include the following:\n> * Active\n> * Reserved\n> * Deprecated\n\nThis resource will retrieve the next available VLAN ID from a given VLAN group (specified by ID).",
		Tags:         true,
		CustomFields: true,
		Put:          true,
		Attrs: []fwresource.Attr{
			{Name: "comments", Kind: fwresource.String, Optional: true, Description: "Free-form comments about the available vlan."},
			{Name: "description", Kind: fwresource.String, Optional: true, Description: "A description of the available vlan."},
			{Name: "group_id", Kind: fwresource.Int, Key: "group", Ref: true, Nullable: true, Optional: true, Description: "The ID of the group associated with the available vlan."},
			{Name: "name", Kind: fwresource.String, Required: true, Description: "The name of the available vlan."},
			{Name: "owner_id", Kind: fwresource.Int, Key: "owner", Ref: true, Nullable: true, Optional: true, Description: "The ID of the owner associated with the available vlan."},
			{Name: "qinq_role", Kind: fwresource.String, Nullable: true, Optional: true, Description: "One of [svlan, cvlan].", OneOf: []string{"cvlan", "svlan"}},
			{Name: "qinq_svlan", Kind: fwresource.Int, Nullable: true, Optional: true, Description: "The qinq svlan of the available vlan."},
			{Name: "role_id", Kind: fwresource.Int, Key: "role", Ref: true, Nullable: true, Optional: true, Description: "The ID of the role associated with the available vlan."},
			{Name: "site_id", Kind: fwresource.Int, Key: "site", Ref: true, Nullable: true, Optional: true, Description: "The ID of the site associated with the available vlan."},
			{Name: "status", Kind: fwresource.String, Optional: true, Computed: true, Description: "The status of the available vlan."},
			{Name: "tenant_id", Kind: fwresource.Int, Key: "tenant", Ref: true, Nullable: true, Optional: true, Description: "The ID of the tenant associated with the available vlan."},
			{Name: "vid", Kind: fwresource.Int, Computed: true, Description: "The VLAN ID (1-4094)."},
		},
		ModifyBody: func(_ context.Context, _, prior fwresource.Values, body map[string]any) error {
			if prior != nil {
				body["vid"] = intValue(prior, "vid")
			}
			return nil
		},
		Allocate: func(ctx context.Context, st *providerstate.State, plan fwresource.Values, body map[string]any) (map[string]any, error) {
			groupID := intValue(plan, "group_id")
			if groupID == 0 {
				return nil, fmt.Errorf("group_id is required to allocate a VLAN")
			}
			data := map[string]any{}
			for k, v := range body {
				if k != "group" && v != nil && v != "" {
					data[k] = v
				}
			}
			path := fmt.Sprintf("ipam/vlan-groups/%d/available-vlans/", groupID)
			created, err := netboxapi.Request[map[string]any](ctx, st.Rest, http.MethodPost, path, nil, data)
			if err != nil {
				return nil, err
			}
			return *created, nil
		},
	}
}
