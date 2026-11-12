package fwresource

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Every kind of resource and data source rejects provider data of the wrong type, and does nothing
// when the provider is not configured yet.
func TestConfigureRejectsWrongProviderData(t *testing.T) {
	ctx := context.Background()
	for name, r := range map[string]resource.ResourceWithConfigure{
		"resource":   New(&Spec{TypeName: "x"})().(resource.ResourceWithConfigure),
		"assign":     NewAssign(AssignSpec{TypeName: "x"})().(resource.ResourceWithConfigure),
		"device tag": NewDeviceTag().(resource.ResourceWithConfigure),
	} {
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: "wrong"}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("%s accepted provider data of the wrong type", name)
		}
		resp = resource.ConfigureResponse{}
		r.Configure(ctx, resource.ConfigureRequest{}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("%s failed without provider data", name)
		}
	}

	ds := NewDataSource(&DataSourceSpec{TypeName: "x"})().(datasource.DataSourceWithConfigure)
	var resp datasource.ConfigureResponse
	ds.Configure(ctx, datasource.ConfigureRequest{ProviderData: "wrong"}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("data source accepted provider data of the wrong type")
	}
	resp = datasource.ConfigureResponse{}
	ds.Configure(ctx, datasource.ConfigureRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Error("data source failed without provider data")
	}
}

func TestUpgradeStateErrors(t *testing.T) {
	ctx := context.Background()
	upgrade := func(fn func(context.Context, map[string]any) (map[string]any, error), raw *tfprotov6.RawState) resource.UpgradeStateResponse {
		r := New(&Spec{TypeName: "x", Upgraders: map[int64]func(context.Context, map[string]any) (map[string]any, error){0: fn}})().(resource.ResourceWithUpgradeState)
		var resp resource.UpgradeStateResponse
		r.UpgradeState(ctx)[0].StateUpgrader(ctx, resource.UpgradeStateRequest{RawState: raw}, &resp)
		return resp
	}
	ok := func(_ context.Context, m map[string]any) (map[string]any, error) { return m, nil }

	if resp := upgrade(ok, &tfprotov6.RawState{JSON: []byte(`{"a":1}`)}); resp.Diagnostics.HasError() || resp.DynamicValue == nil {
		t.Errorf("a valid state failed: %v", resp.Diagnostics)
	}
	if resp := upgrade(ok, nil); !resp.Diagnostics.HasError() {
		t.Error("a state without JSON was accepted")
	}
	if resp := upgrade(ok, &tfprotov6.RawState{JSON: []byte(`not json`)}); !resp.Diagnostics.HasError() {
		t.Error("invalid JSON was accepted")
	}
	fail := func(context.Context, map[string]any) (map[string]any, error) { return nil, errors.New("boom") }
	if resp := upgrade(fail, &tfprotov6.RawState{JSON: []byte(`{}`)}); !resp.Diagnostics.HasError() {
		t.Error("a failing upgrader was accepted")
	}
	unencodable := func(context.Context, map[string]any) (map[string]any, error) {
		return map[string]any{"c": make(chan int)}, nil
	}
	if resp := upgrade(unencodable, &tfprotov6.RawState{JSON: []byte(`{}`)}); !resp.Diagnostics.HasError() {
		t.Error("a state that cannot be encoded was accepted")
	}
}

func TestDeviceTagUpdateDoesNothing(t *testing.T) {
	var resp resource.UpdateResponse
	NewDeviceTag().Update(context.Background(), resource.UpdateRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Error("update failed")
	}
}
