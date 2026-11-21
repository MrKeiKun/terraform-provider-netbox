package fwprovider

import (
	"context"
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func specOf(t *testing.T, name string) *fwresource.Spec {
	t.Helper()
	for _, s := range standardResources {
		if s.TypeName == name {
			return s
		}
	}
	t.Fatalf("no spec for %s", name)
	return nil
}

// NetBox never returns a password, so a user needs one when it is created and not afterwards.
func TestUserNeedsAPasswordOnlyToCreate(t *testing.T) {
	s := specOf(t, "netbox_user")
	plan := fwresource.Values{"password": types.StringNull()}
	if err := s.ModifyBody(context.Background(), plan, nil, map[string]any{}); err == nil {
		t.Error("a user without a password was accepted on create")
	}
	if err := s.ModifyBody(context.Background(), plan, fwresource.Values{}, map[string]any{}); err != nil {
		t.Errorf("an update without a password failed: %v", err)
	}
	plan["password"] = types.StringValue("secret")
	if err := s.ModifyBody(context.Background(), plan, nil, map[string]any{}); err != nil {
		t.Errorf("a user with a password failed: %v", err)
	}
}

// A rack with a rack type gets its dimensions from the type, so the state does not hold them and a
// configuration written from the state does not conflict with rack_type_id.
func TestRackWithRackTypeDoesNotKeepDerivedDimensions(t *testing.T) {
	s := specOf(t, "netbox_rack")
	out := fwresource.Values{"form_factor": types.StringValue("4-post-cabinet"), "width": types.Int64Value(19), "u_height": types.Int64Value(42)}
	if err := s.ModifyState(context.Background(), map[string]any{"rack_type": map[string]any{"id": float64(1)}}, nil, out); err != nil {
		t.Fatal(err)
	}
	if !out["form_factor"].IsNull() || !out["width"].IsNull() || !out["u_height"].IsNull() {
		t.Errorf("derived dimensions stayed in the state: %v", out)
	}

	out = fwresource.Values{"form_factor": types.StringValue("4-post-cabinet"), "width": types.Int64Value(19), "u_height": types.Int64Value(42)}
	_ = s.ModifyState(context.Background(), map[string]any{"rack_type": nil}, nil, out)
	if out["u_height"].IsNull() {
		t.Error("a rack without a rack type lost its height")
	}
}

// The values of these attributes come back from NetBox, so leaving them out of the configuration
// must not clear them.
func TestSecretsAreKeptWhenNotConfigured(t *testing.T) {
	for typeName, attr := range map[string]string{
		"netbox_ike_policy":    "preshared_key",
		"netbox_fhrp_group":    "auth_key",
		"netbox_webhook":       "additional_headers",
		"netbox_wireless_lan":  "auth_psk",
		"netbox_wireless_link": "auth_psk",
		"netbox_data_source":   "parameters",
	} {
		found := false
		for _, a := range specOf(t, typeName).Attrs {
			if a.Name == attr {
				found = true
				if !a.Computed || !a.Sensitive {
					t.Errorf("%s.%s: computed=%v sensitive=%v, want both", typeName, attr, a.Computed, a.Sensitive)
				}
			}
		}
		if !found {
			t.Errorf("%s has no %s", typeName, attr)
		}
	}
}

// NetBox derives the data source, the data path and the automatic sync from the data file, so the plan
// has to predict that they are cleared when the data file is taken away.
func TestDataFilePlanClearsWhatNetBoxDerives(t *testing.T) {
	plan := fwresource.Values{
		"data_source_id":    types.Int64Value(5),
		"data_path":         types.StringValue("t.j2"),
		"auto_sync_enabled": types.BoolValue(true),
	}
	prior := fwresource.Values{"data_file_id": types.Int64Value(1)}

	got := dataFilePlan(fwresource.Values{"data_file_id": types.Int64Null()}, plan, prior)
	if got["data_source_id"].String() != "0" || got["data_path"].String() != `""` || got["auto_sync_enabled"].String() != "false" {
		t.Errorf("unlinking did not clear the derived values: %v", got)
	}
	if got := dataFilePlan(fwresource.Values{"data_file_id": types.Int64Value(1)}, plan, prior); got != nil {
		t.Errorf("a linked object was changed: %v", got)
	}
	if got := dataFilePlan(fwresource.Values{"data_file_id": types.Int64Null()}, plan, nil); got != nil {
		t.Errorf("a new object was changed: %v", got)
	}
	// a data source without a data file (older NetBox versions keep one) is left alone
	if got := dataFilePlan(fwresource.Values{"data_file_id": types.Int64Null()}, plan, fwresource.Values{"data_file_id": types.Int64Value(0)}); got != nil {
		t.Errorf("an object that never had a data file was changed: %v", got)
	}
}

// An old attribute name and its new name stay equal in the plan: the one that is left out takes the
// value of the other.
func TestAliasPlanFollowsTheNameThatIsWritten(t *testing.T) {
	set := portSet("dcim.site")
	null := types.SetNull(types.StringType)
	plan := fwresource.Values{"content_types": set, "object_types": types.SetUnknown(types.StringType)}
	got := aliasPlan("content_types", "object_types")(fwresource.Values{"content_types": set, "object_types": null}, plan, nil)
	if got["object_types"] == nil || !got["object_types"].Equal(set) {
		t.Errorf("the new name did not follow the old one: %v", got)
	}
	plan = fwresource.Values{"content_types": types.SetUnknown(types.StringType), "object_types": set}
	got = aliasPlan("content_types", "object_types")(fwresource.Values{"content_types": null, "object_types": set}, plan, nil)
	if got["content_types"] == nil || !got["content_types"].Equal(set) {
		t.Errorf("the old name did not follow the new one: %v", got)
	}
	if got := aliasPlan("content_types", "object_types")(fwresource.Values{"content_types": set, "object_types": set}, plan, nil); got != nil {
		t.Errorf("both written, nothing to follow: %v", got)
	}
}

func TestDataFileBodyLeavesOutANullDataFile(t *testing.T) {
	create := map[string]any{"data_file": nil, "name": "a"}
	if err := dataFileBody(context.Background(), nil, nil, create); err != nil {
		t.Fatal(err)
	}
	if _, ok := create["data_file"]; ok {
		t.Fatalf("a create must not send data_file: null, got %v", create)
	}

	hadFile := fwresource.Values{"data_file_id": types.Int64Value(7)}
	update := map[string]any{"data_file": nil}
	if err := dataFileBody(context.Background(), nil, hadFile, update); err != nil {
		t.Fatal(err)
	}
	if _, ok := update["data_file"]; !ok {
		t.Fatal("taking a data file away must still send null")
	}
}

func TestUserKeepsStaffWhenNetBoxHasNoIsStaff(t *testing.T) {
	s := specOf(t, "netbox_user")
	prior := fwresource.Values{"staff": types.BoolValue(true)}
	out := fwresource.Values{"staff": types.BoolValue(false)}
	if err := s.ModifyState(context.Background(), map[string]any{"username": "a"}, prior, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["staff"].(types.Bool); !v.ValueBool() {
		t.Error("staff = true was not kept when NetBox returns no is_staff")
	}
	out = fwresource.Values{"staff": types.BoolValue(false)}
	if err := s.ModifyState(context.Background(), map[string]any{"is_staff": false}, prior, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["staff"].(types.Bool); v.ValueBool() {
		t.Error("the value NetBox returns was overwritten")
	}
}

func TestDataFileBodyKeepsOtherValues(t *testing.T) {
	body := map[string]any{"data_file": float64(5)}
	if err := dataFileBody(context.Background(), nil, nil, body); err != nil {
		t.Fatal(err)
	}
	if body["data_file"] != float64(5) {
		t.Errorf("a data file was changed: %v", body)
	}
	empty := map[string]any{"name": "a"}
	if err := dataFileBody(context.Background(), nil, nil, empty); err != nil || len(empty) != 1 {
		t.Errorf("a body without data_file was changed: %v", empty)
	}
}

func TestDataFileStateKeepsDataSourceAtZeroWithoutOne(t *testing.T) {
	out := fwresource.Values{"data_source_id": types.Int64Value(3)}
	if err := dataFileState(context.Background(), map[string]any{"data_source": nil}, nil, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["data_source_id"].(types.Int64); v.ValueInt64() != 0 {
		t.Errorf("data_source_id = %v, want 0", out["data_source_id"])
	}
	out = fwresource.Values{"data_source_id": types.Int64Value(3)}
	if err := dataFileState(context.Background(), map[string]any{"data_source": map[string]any{"id": 3}}, nil, out); err != nil {
		t.Fatal(err)
	}
	if v, _ := out["data_source_id"].(types.Int64); v.ValueInt64() != 3 {
		t.Errorf("a data source was dropped: %v", out["data_source_id"])
	}
}
