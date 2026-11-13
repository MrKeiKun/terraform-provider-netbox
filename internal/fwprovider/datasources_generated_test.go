package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// A generated plural data source reads without a filter, also when nothing is found.
func TestGeneratedPluralDataSourcesRead(t *testing.T) {
	_, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)

	count := 0
	for _, spec := range generatedDataSources {
		if spec.Plural == nil || spec.SinceVersion != "" {
			continue // the fake NetBox runs 4.4.0, which does not have the newer object types
		}
		count++
		typ := h.schemas.DataSourceSchemas[spec.TypeName].ValueType()
		resp, err := h.server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{
			TypeName: spec.TypeName,
			Config:   h.dynamic(h.object(typ, map[string]tftypes.Value{})),
		})
		if err != nil {
			t.Fatalf("%s: %v", spec.TypeName, err)
		}
		h.noErrors(spec.TypeName, resp.Diagnostics)
	}
	if count == 0 {
		t.Fatal("no generated plural data sources")
	}
}
