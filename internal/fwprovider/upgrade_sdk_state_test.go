package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// plain turns a value of the state into plain Go values, with sets and lists as sorted slices.
func plain(t *testing.T, v tftypes.Value) any {
	t.Helper()
	if v.IsNull() {
		return nil
	}
	switch {
	case v.Type().Is(tftypes.String):
		var s string
		_ = v.As(&s)
		return s
	case v.Type().Is(tftypes.Number):
		var n big.Float
		_ = v.As(&n)
		f, _ := n.Float64()
		return f
	case v.Type().Is(tftypes.Bool):
		var b bool
		_ = v.As(&b)
		return b
	case v.Type().Is(tftypes.Map{}) || v.Type().Is(tftypes.Object{}):
		var m map[string]tftypes.Value
		_ = v.As(&m)
		out := map[string]any{}
		for k, e := range m {
			out[k] = plain(t, e)
		}
		return out
	default:
		var l []tftypes.Value
		_ = v.As(&l)
		out := make([]any, 0, len(l))
		for _, e := range l {
			out = append(out, plain(t, e))
		}
		sort.Slice(out, func(i, j int) bool { return fmt.Sprint(out[i]) < fmt.Sprint(out[j]) })
		return out
	}
}

func sorted(v any) any {
	switch x := v.(type) {
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = sorted(e)
		}
		sort.Slice(out, func(i, j int) bool { return fmt.Sprint(out[i]) < fmt.Sprint(out[j]) })
		return out
	case map[string]any:
		out := map[string]any{}
		for k, e := range x {
			out[k] = sorted(e)
		}
		return out
	}
	return v
}

// State written by the provider that was built on the plugin SDK must load unchanged: the same
// attributes with the same values, nothing dropped and no error.
func TestUpgradeStateWrittenBySDKProvider(t *testing.T) {
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatal(err)
	}
	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}

	for typeName, state := range map[string]string{
		"netbox_site": `{"id":"4","name":"dc1","slug":"dc1","status":"active","description":"main","tags":["blue","red"],"tags_all":["blue","red","managed"],
			"custom_fields":{"owner":"net"},"latitude":52.5,"longitude":13.4,"asn_ids":[1,2],"group_id":0}`,
		"netbox_tenant":               `{"id":"7","name":"acme","slug":"acme","description":"","group_id":null,"tags":[],"tags_all":[]}`,
		"netbox_prefix":               `{"id":"9","prefix":"10.0.0.0/24","status":"active","is_pool":true,"mark_utilized":false,"vrf_id":0,"tags":["x"],"tags_all":["x"]}`,
		"netbox_ip_address":           `{"id":"11","ip_address":"10.0.0.5/24","status":"active","dns_name":"a.example.com","object_type":"dcim.interface","interface_id":3,"tags":[],"tags_all":[]}`,
		"netbox_vlan":                 `{"id":"12","name":"v10","vid":10,"status":"active","tags":[],"tags_all":[]}`,
		"netbox_tag":                  `{"id":"13","name":"blue","slug":"blue","color_hex":"0000ff","description":""}`,
		"netbox_interface":            `{"id":"14","name":"eth0","virtual_machine_id":2,"enabled":true,"mtu":1500,"tagged_vlans":[5,6],"tags":[],"tags_all":[]}`,
		"netbox_device":               `{"id":"15","name":"sw1","device_type_id":1,"role_id":2,"site_id":3,"status":"active","local_context_data":"{\"a\":1}","tags":["x"],"tags_all":["x"],"custom_fields":{}}`,
		"netbox_custom_field":         `{"id":"16","name":"owner","type":"text","content_types":["dcim.site"],"weight":100,"required":false}`,
		"netbox_device_front_port":    `{"id":"17","name":"fp1","device_id":1,"type":"8p8c","rear_ports":[{"rear_port_id":4,"position":1,"rear_port_position":1}],"tags":[],"tags_all":[]}`,
		"netbox_available_ip_address": `{"id":"18","prefix_id":3,"status":"active","ip_address":"10.0.0.7/24","tags":[],"tags_all":[]}`,
		"netbox_circuit":              `{"id":"19","cid":"c1","provider_id":1,"type_id":2,"status":"active","tags":[],"tags_all":[]}`,
	} {
		resp, err := server.UpgradeResourceState(context.Background(), &tfprotov6.UpgradeResourceStateRequest{
			TypeName: typeName,
			Version:  0,
			RawState: &tfprotov6.RawState{JSON: []byte(state)},
		})
		if err != nil {
			t.Fatalf("%s: %v", typeName, err)
		}
		for _, d := range resp.Diagnostics {
			t.Errorf("%s: %s: %s", typeName, d.Summary, d.Detail)
		}
		if resp.UpgradedState == nil {
			t.Errorf("%s: no upgraded state", typeName)
			continue
		}
		sch := schemas.ResourceSchemas[typeName]
		if sch == nil {
			t.Errorf("%s is not a resource", typeName)
			continue
		}
		value, err := resp.UpgradedState.Unmarshal(sch.ValueType())
		if err != nil {
			t.Errorf("%s: %v", typeName, err)
			continue
		}
		got := plain(t, value).(map[string]any)

		var want map[string]any
		if err := json.Unmarshal([]byte(state), &want); err != nil {
			t.Fatal(err)
		}
		for k, w := range want {
			if w == nil {
				continue
			}
			if !reflect.DeepEqual(sorted(got[k]), sorted(w)) {
				t.Errorf("%s.%s: state has %v, was %v", typeName, k, got[k], w)
			}
		}
	}
}
