package fwprovider

import (
	"context"
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func standardSpec(t *testing.T, name string) *fwresource.Spec {
	t.Helper()
	for _, s := range standardResources {
		if s.TypeName == name {
			return s
		}
	}
	t.Fatalf("no resource %s", name)
	return nil
}

func mustPanic(t *testing.T, what string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s did not panic", what)
		}
	}()
	f()
}

func TestEventRuleSetsTheObjectType(t *testing.T) {
	s := standardSpec(t, "netbox_event_rule")
	for action, want := range map[string]string{"webhook": "extras.webhook", "script": "extras.script"} {
		body := map[string]any{}
		if err := s.ModifyBody(context.Background(), fwresource.Values{"action_type": types.StringValue(action)}, nil, body); err != nil {
			t.Fatal(err)
		}
		if body["action_object_type"] != want {
			t.Errorf("%s: %v", action, body["action_object_type"])
		}
	}
}

func TestContactGroupCompatibility(t *testing.T) {
	s := standardSpec(t, "netbox_contact")
	ctx := context.Background()

	body := map[string]any{}
	_ = s.ModifyBody(ctx, fwresource.Values{"group_id": types.Int64Value(3), "group_ids": types.SetNull(types.Int64Type)}, nil, body)
	if groups, _ := body["groups"].([]any); len(groups) != 1 || groups[0] != int64(3) {
		t.Errorf("groups = %v", body["groups"])
	}

	// the single group is read back from the list of groups
	out := fwresource.Values{}
	prior := fwresource.Values{"group_id": types.Int64Value(3), "group_ids": types.SetValueMust(types.Int64Type, []attr.Value{})}
	obj := map[string]any{"groups": []any{map[string]any{"id": float64(3)}}}
	if err := s.ModifyState(ctx, obj, prior, out); err != nil {
		t.Fatal(err)
	}
	if out["group_id"] != types.Int64Value(3) {
		t.Errorf("group_id = %v", out["group_id"])
	}

	// nothing is set when the group was not configured
	out = fwresource.Values{}
	_ = s.ModifyState(ctx, obj, fwresource.Values{"group_id": types.Int64Value(0)}, out)
	if out["group_id"] != types.Int64Value(0) {
		t.Errorf("group_id = %v", out["group_id"])
	}
}

func TestTokenKeepsTheKnownKey(t *testing.T) {
	s := standardSpec(t, "netbox_token")
	ctx := context.Background()

	out := fwresource.Values{}
	_ = s.ModifyState(ctx, map[string]any{"key": "shown"}, fwresource.Values{}, out)
	if len(out) != 0 {
		t.Errorf("a key that NetBox shows was replaced: %v", out)
	}
	out = fwresource.Values{}
	_ = s.ModifyState(ctx, map[string]any{}, fwresource.Values{"key": types.StringValue("known")}, out)
	if out["key"] != types.StringValue("known") {
		t.Errorf("key = %v", out["key"])
	}
	out = fwresource.Values{}
	_ = s.ModifyState(ctx, map[string]any{}, fwresource.Values{"key": types.StringNull()}, out)
	if !out["key"].IsNull() {
		t.Errorf("key = %v", out["key"])
	}
}

func TestPortMappingWithoutPriorValues(t *testing.T) {
	s := &fwresource.Spec{}
	portMappingHooks(s, "rear_ports", "rear_port", true)
	body := map[string]any{"rear_port": int64(1)}
	// nothing is known about the list and the single port, so nothing is the same as before
	if err := s.ModifyBody(context.Background(), fwresource.Values{}, fwresource.Values{"rear_ports": types.ListNull(types.StringType)}, body); err != nil {
		t.Fatal(err)
	}
	if err := s.ModifyBody(context.Background(), fwresource.Values{}, fwresource.Values{}, body); err != nil {
		t.Fatal(err)
	}
}

func TestSpecHelpersPanicOnMistakes(t *testing.T) {
	s := &fwresource.Spec{TypeName: "x", Attrs: []fwresource.Attr{{Name: "n", Kind: fwresource.NestedList, Sub: []fwresource.Attr{{Name: "a"}}}}}
	mustPanic(t, "setAttr", func() { setAttr(s, "missing", func(*fwresource.Attr) {}) })
	mustPanic(t, "removeAttr", func() { removeAttr(s, "missing") })
	mustPanic(t, "setSub on a missing attribute", func() { setSub(s, "missing", "a", func(*fwresource.Attr) {}) })
	mustPanic(t, "setSub on a missing sub attribute", func() { setSub(s, "n", "missing", func(*fwresource.Attr) {}) })
	mustPanic(t, "setDSAttr", func() { setDSAttr(&fwresource.DataSourceSpec{TypeName: "x"}, "missing", func(*fwresource.Attr) {}) })
	mustPanic(t, "derive", func() { derive("netbox_nothing") })

	// rename also changes the descriptions of nested attributes
	s.Attrs[0].Sub[0].Description = "the prefix"
	rename(s, "the prefix", "the available prefix")
	if s.Attrs[0].Sub[0].Description != "the available prefix" {
		t.Errorf("description = %q", s.Attrs[0].Sub[0].Description)
	}
}

func TestDeviceAfterWriteDoesNothingWithoutAChange(t *testing.T) {
	s := device()
	plan := fwresource.Values{"virtual_chassis_id": types.Int64Value(1), "virtual_chassis_master": types.BoolValue(false)}
	if changed, err := s.AfterWrite(context.Background(), nil, 1, plan, nil); changed || err != nil {
		t.Errorf("create without master: %v %v", changed, err)
	}
	if changed, err := s.AfterWrite(context.Background(), nil, 1, plan, plan); changed || err != nil {
		t.Errorf("update without a change: %v %v", changed, err)
	}
	if changed, err := s.AfterWrite(context.Background(), nil, 1, fwresource.Values{"virtual_chassis_id": types.Int64Value(0)}, nil); changed || err != nil {
		t.Errorf("no chassis: %v %v", changed, err)
	}
}
