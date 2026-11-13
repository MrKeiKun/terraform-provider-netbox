package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// No two resources and no two data sources have the same name.
func TestTypeNamesAreUnique(t *testing.T) {
	resources, dataSources := Inventory()
	seen := map[string]bool{}
	for _, r := range resources {
		if seen[r.TypeName] {
			t.Errorf("resource %s is defined twice", r.TypeName)
		}
		seen[r.TypeName] = true
	}
	seen = map[string]bool{}
	for _, d := range dataSources {
		if seen[d.TypeName] {
			t.Errorf("data source %s is defined twice", d.TypeName)
		}
		seen[d.TypeName] = true
	}
}

// A data source with a second name has the schema of the data source it copies.
func TestDataSourceAliases(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)

	for alias, original := range dataSourceAliases {
		a, o := h.schemas.DataSourceSchemas[alias], h.schemas.DataSourceSchemas[original]
		if a == nil || o == nil {
			t.Errorf("%s or %s is not a data source", alias, original)
			continue
		}
		if !a.ValueType().Equal(o.ValueType()) {
			t.Errorf("%s does not have the attributes of %s", alias, original)
		}
	}
}

func TestSupportedVersionsDataSource(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)

	typ := h.schemas.DataSourceSchemas["netbox_supported_versions"].ValueType()
	resp, err := h.server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{
		TypeName: "netbox_supported_versions",
		Config:   h.dynamic(h.object(typ, map[string]tftypes.Value{})),
	})
	if err != nil {
		t.Fatal(err)
	}
	h.noErrors("read", resp.Diagnostics)
	var versions []tftypes.Value
	if err := h.decode(typ, resp.State)["versions"].As(&versions); err != nil || len(versions) == 0 {
		t.Errorf("no versions: %v %v", versions, err)
	}
}
