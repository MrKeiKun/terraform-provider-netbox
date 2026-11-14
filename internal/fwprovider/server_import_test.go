package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestImportByName(t *testing.T) {
	fake, srv := newFakeNetBox()
	defer srv.Close()
	h := newHarness(t, srv.URL)
	ctx := context.Background()

	fake.put("dcim/sites", map[string]any{"id": float64(3), "name": "Paris", "slug": "paris"})
	fake.put("dcim/sites", map[string]any{"id": float64(4), "name": "Berlin", "slug": "berlin"})
	fake.put("extras/tags", map[string]any{"id": float64(8), "name": "blue", "slug": "blue-1"})
	fake.put("extras/tags", map[string]any{"id": float64(9), "name": "blue", "slug": "blue-2"})

	importID := func(typeName, id string) (string, bool) {
		resp, err := h.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: typeName, ID: id})
		if err != nil {
			t.Fatal(err)
		}
		if h.hasError(resp.Diagnostics) {
			return "", false
		}
		state := h.decode(h.schemas.ResourceSchemas[typeName].ValueType(), resp.ImportedResources[0].State)
		var got string
		if err := state["id"].As(&got); err != nil {
			t.Fatal(err)
		}
		return got, true
	}

	if id, ok := importID("netbox_site", "berlin"); !ok || id != "4" {
		t.Errorf("import by slug gave %q, %v", id, ok)
	}
	if id, ok := importID("netbox_site", "3"); !ok || id != "3" {
		t.Errorf("import by number gave %q, %v", id, ok)
	}
	if _, ok := importID("netbox_site", "nowhere"); ok {
		t.Error("a slug that does not exist was imported")
	}
	if id, ok := importID("netbox_tag", "blue-2"); !ok || id != "9" {
		t.Errorf("import of a tag by slug gave %q, %v", id, ok)
	}
}
