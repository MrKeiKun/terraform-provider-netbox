package fwresource

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// readCustomPlural reads a plural data source that looks up its objects itself.
func readCustomPlural(t *testing.T, objs []map[string]any, fail error) ([]map[string]tftypes.Value, *datasource.ReadResponse) {
	spec := &DataSourceSpec{
		TypeName: "netbox_things",
		Attrs: []Attr{
			{Name: "name", Kind: String, Required: true},
			{Name: "size", Kind: Int},
		},
		Plural: &Plural{Result: "things", What: "things", NoFilter: true},
		Custom: func(context.Context, *providerstate.State, Values, datasource.ReadRequest) ([]map[string]any, error) {
			return objs, fail
		},
	}
	d := NewDataSource(spec)().(*DataSource)
	d.state = &providerstate.State{}

	ctx := context.Background()
	schema := d.spec.buildSchema()
	typ := schema.Type().TerraformType(ctx).(tftypes.Object)
	nulls := map[string]tftypes.Value{}
	for name, at := range typ.AttributeTypes {
		nulls[name] = tftypes.NewValue(at, nil)
	}
	raw := tftypes.NewValue(typ, nulls)
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schema, Raw: raw}}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: schema, Raw: raw}}, resp)
	if resp.Diagnostics.HasError() {
		return nil, resp
	}

	var state map[string]tftypes.Value
	if err := resp.State.Raw.As(&state); err != nil {
		t.Fatal(err)
	}
	var list []tftypes.Value
	if err := state["things"].As(&list); err != nil {
		t.Fatal(err)
	}
	out := make([]map[string]tftypes.Value, len(list))
	for i, v := range list {
		if err := v.As(&out[i]); err != nil {
			t.Fatal(err)
		}
	}
	return out, resp
}

func TestPluralDataSourceWithItsOwnLookup(t *testing.T) {
	things, _ := readCustomPlural(t, []map[string]any{
		{"name": "a", "size": float64(3)},
		// an object that lacks a field still gives a result, with the field empty
		{"size": float64(4)},
	}, nil)
	if len(things) != 2 {
		t.Fatalf("got %d things, want 2", len(things))
	}
	var name string
	if err := things[0]["name"].As(&name); err != nil || name != "a" {
		t.Errorf("name of the first thing: %q, %v", name, err)
	}
	var size big.Float
	if err := things[1]["size"].As(&size); err != nil || size.Cmp(big.NewFloat(4)) != 0 {
		t.Errorf("size of the second thing: %s, %v", size.String(), err)
	}
	if !things[1]["name"].IsNull() {
		var got string
		_ = things[1]["name"].As(&got)
		if got != "" {
			t.Errorf("the missing name is %q", got)
		}
	}

	// no objects is an empty list, not null
	empty, resp := readCustomPlural(t, nil, nil)
	if resp.Diagnostics.HasError() || len(empty) != 0 {
		t.Errorf("no objects: %d things, %v", len(empty), resp.Diagnostics)
	}
	var state map[string]tftypes.Value
	_ = resp.State.Raw.As(&state)
	if state["things"].IsNull() {
		t.Errorf("the result is null when there are no objects")
	}

	// an error of the lookup is reported
	_, resp = readCustomPlural(t, nil, errors.New("lookup failed"))
	if !resp.Diagnostics.HasError() {
		t.Errorf("no error when the lookup fails")
	}
}
