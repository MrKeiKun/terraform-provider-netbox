package fwprovider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// The fake NetBox reports version 4.4.0. An attribute that needs a later version is an error in the plan,
// and one that the version has is not.
func TestAttributeNeedsNewerNetBox(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)

	plan := func(typeName string, set map[string]tftypes.Value) []*tfprotov6.Diagnostic {
		typ := h.schemas.ResourceSchemas[typeName].ValueType()
		cfg := h.object(typ, set)
		prior := tftypes.NewValue(typ, nil)
		resp, err := h.server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{
			TypeName:         typeName,
			PriorState:       h.dynamic(prior),
			ProposedNewState: h.dynamic(h.proposed(typeName, typ, prior, cfg)),
			Config:           h.dynamic(cfg),
		})
		if err != nil {
			t.Fatal(err)
		}
		return resp.Diagnostics
	}

	base := map[string]tftypes.Value{"name": str("site"), "slug": str("site")}
	if diags := plan("netbox_site", base); h.hasError(diags) {
		t.Errorf("a site without owner_id was rejected: %v", diags)
	}

	withOwner := map[string]tftypes.Value{"name": str("site"), "slug": str("site"), "owner_id": num(1)}
	diags := plan("netbox_site", withOwner)
	if !h.hasError(diags) {
		t.Fatal("owner_id was accepted on NetBox 4.4.0")
	}
	if !strings.Contains(diags[0].Detail, "4.5.0") {
		t.Errorf("the error does not say which version is needed: %s", diags[0].Detail)
	}

	// a whole object type that the server does not have
	diags = plan("netbox_cooling_source", map[string]tftypes.Value{"name": str("c")})
	if !h.hasError(diags) {
		t.Error("netbox_cooling_source was accepted on NetBox 4.4.0")
	}
}

// A data source, and a filter of one, that the NetBox server is too old for are an error.
func TestDataSourceNeedsNewerNetBox(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)

	read := func(typeName string, set map[string]tftypes.Value) []*tfprotov6.Diagnostic {
		typ := h.schemas.DataSourceSchemas[typeName].ValueType()
		resp, err := h.server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{
			TypeName: typeName,
			Config:   h.dynamic(h.object(typ, set)),
		})
		if err != nil {
			t.Fatal(err)
		}
		return resp.Diagnostics
	}

	if diags := read("netbox_cooling_sources", nil); !h.hasError(diags) {
		t.Error("netbox_cooling_sources was read on NetBox 4.4.0")
	}

	filter := tftypes.NewValue(
		tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String, "value": tftypes.String}}},
		[]tftypes.Value{tftypes.NewValue(
			tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String, "value": tftypes.String}},
			map[string]tftypes.Value{"name": str("owner_id"), "value": str("1")},
		)},
	)
	diags := read("netbox_sites", map[string]tftypes.Value{"filter": filter})
	if !h.hasError(diags) {
		t.Fatal("the owner_id filter was accepted on NetBox 4.4.0")
	}
	if !strings.Contains(diags[0].Detail, "owner_id") {
		t.Errorf("the error does not name the filter: %s", diags[0].Detail)
	}
	if diags := read("netbox_sites", nil); h.hasError(diags) {
		t.Errorf("a plain read of netbox_sites failed: %v", diags)
	}
}
