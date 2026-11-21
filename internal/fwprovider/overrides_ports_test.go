package fwprovider

import (
	"context"
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func portSet(vals ...string) types.Set {
	elems := make([]attr.Value, len(vals))
	for i, v := range vals {
		elems[i] = types.StringValue(v)
	}
	return types.SetValueMust(types.StringType, elems)
}

// NetBox rejects an update with both port_mappings and the older protocol and ports.
func TestLegacyPortsOrMappingsSendsOneFamily(t *testing.T) {
	ctx := context.Background()
	body := func() map[string]any {
		return map[string]any{"protocol": "tcp", "ports": []any{81}, "port_mappings": []any{"tcp/80"}}
	}

	// a new object keeps whatever was configured
	b := body()
	if err := legacyPortsOrMappings(ctx, fwresource.Values{}, nil, b); err != nil || len(b) != 3 {
		t.Errorf("create changed the body: %v %v", b, err)
	}

	// ports changed, port_mappings still holds the value of the state: send the ports
	prior := fwresource.Values{"port_mappings": portSet("tcp/80")}
	b = body()
	_ = legacyPortsOrMappings(ctx, fwresource.Values{"port_mappings": portSet("tcp/80")}, prior, b)
	if _, has := b["port_mappings"]; has || b["ports"] == nil {
		t.Errorf("want ports only, got %v", b)
	}

	// port_mappings not known yet: the same
	b = body()
	_ = legacyPortsOrMappings(ctx, fwresource.Values{"port_mappings": types.SetUnknown(types.StringType)}, prior, b)
	if _, has := b["port_mappings"]; has || b["ports"] == nil {
		t.Errorf("want ports only while port_mappings is unknown, got %v", b)
	}

	// port_mappings changed: send the mappings and leave the older fields out
	b = body()
	_ = legacyPortsOrMappings(ctx, fwresource.Values{"port_mappings": portSet("udp/53")}, prior, b)
	if b["port_mappings"] == nil || b["ports"] != nil || b["protocol"] != nil {
		t.Errorf("want port_mappings only, got %v", b)
	}
}

func TestLegacyPortsPlanMarksMappingsUnknown(t *testing.T) {
	prior := fwresource.Values{"protocol": types.StringValue("tcp"), "ports": types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(80)})}
	same := fwresource.Values{"protocol": types.StringValue("tcp"), "ports": types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(80)})}
	changed := fwresource.Values{"protocol": types.StringValue("tcp"), "ports": types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(81)})}
	unset := fwresource.Values{"port_mappings": types.SetNull(types.StringType)}
	set := fwresource.Values{"port_mappings": portSet("tcp/81")}

	if got := legacyPortsPlan(unset, changed, prior); got["port_mappings"] == nil || !got["port_mappings"].IsUnknown() {
		t.Errorf("a change of ports did not mark port_mappings unknown: %v", got)
	}
	if got := legacyPortsPlan(unset, same, prior); got != nil {
		t.Errorf("an unchanged plan was modified: %v", got)
	}
	if got := legacyPortsPlan(set, changed, prior); got != nil {
		t.Errorf("configured port_mappings were replaced: %v", got)
	}
	if got := legacyPortsPlan(unset, changed, nil); got != nil {
		t.Errorf("a new object was modified: %v", got)
	}
}
